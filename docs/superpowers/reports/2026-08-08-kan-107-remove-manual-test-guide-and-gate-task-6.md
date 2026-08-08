# Task 6 report: Delete `docs/manual-test/` and sweep the remaining contract prose

## Status: DONE

## What was done

**Step 1 — deleted the directory.** `git rm -r docs/manual-test/` staged the removal of all
fourteen guides. This is a staged deletion in the index, not a commit, per the myflow no-commits
rule for this task.

**Step 2 — swept the contract prose** in `skills/myflow-contracts/`:
- `pipeline.md:30-38` — the `IN_PROGRESS` gate paragraph and its table row now say the staged
  diff awaits review "alongside the run instructions the handoff printed," and the table's
  `IN_PROGRESS` "Means" cell reads "the handoff printed the run instructions" instead of "a test
  guide exists."
- `pipeline.md:240-244` (renumbered after the edit) — the "Link, never paste" bullet now names
  only "Diffs and plans"; the following "Every path is absolute" bullet dropped "in generated
  guides."
- `pipeline.md:316` (IntelliJ table) — the `IN_PROGRESS` row now names only "apply worktree
  root," dropping "plus the test guide's absolute path."
- `finish-contract.md` — already clean; the two-commit sentence at lines 125-129 names
  `openspec/` and the preserved `docs/superpowers/` records with no manual-test-guide mention (an
  earlier task had already landed this).
- `plan-provenance.md:279-286` — the quoting-exception table's prose and row now read "this
  document and the change's delta spec" / "this file and the delta spec," dropping the
  manual-test-guide leg.
- `workspace-isolation.md:171` — "The ports actually bound are written into the change's manual
  test guide" became "The ports actually bound are printed in the `/myflow-do` handoff."
- `project-configuration.md:197` — "a run writing the change's manual test guide, most of all"
  became "a run resolving the run instructions, most of all."

**Step 3 — fixed the two stale comments and the fixture path:**
- `scripts/check-references.sh:414` — comment's example shape swapped from
  `docs/manual-test/<name>.md` to `openspec/changes/<name>/tasks.md`.
- `scripts/check-plan-provenance.py` (originally line 588, now ~586-596 after the fix round
  below) — comment's example shape swapped away from the deleted guide's fixture line to the
  archived design doc's demonstration line.
- `scripts/test-check-references.sh:111` — fixture path swapped from `docs/manual-test/<name>.md`
  to `openspec/changes/<name>/tasks.md`, keeping the "unresolvable path skipped" test's shape.

**skills/README.md:60-61** — the `/myflow-do` row's "leaving a manual test guide beside a staged
diff" became "printing the run instructions in its handoff beside a staged diff," and the
following gate row's "run the apps against the guide" became "run the apps."

**Five controller-added sites**, all confirmed by grep and fixed: `skills/myflow-do/SKILL.md:3`
(frontmatter), `:8` (intro sentence), `skills/myflow-status/SKILL.md:112` (example row), `:158`
(next-command mapping), and `skills/myflow-do/SKILL-rationale.md:21` (the brief cited line 79,
which had already been retitled by an earlier task — the actual leftover, found by grepping the
whole file, was at line 21: "section 6 writes the guide's URLs from it").

**One extra site found during the Step 4 sweep, not in the brief or the controller's list:**
`skills/myflow-contracts/pipeline-rationale.md:133` — "`/myflow-do` hands off a manual test guide
and the operator runs the applications against it" → "`/myflow-do`'s handoff prints the run
instructions and the operator runs the applications against them."

**Step 4 — verified the sweep.** Final grep (`grep -rln "manual test guide\|manual-test" skills/
scripts/ docs/ commands/ commands-claude/ rules/ | grep -v 'docs/superpowers/'`) leaves only:
`scripts/check-vocabulary.sh` (retired-command-name allowlist, must stay per the global
constraints), `scripts/test-check-unfinished-work.sh` (Tasks 1-5's own test, not in this task's
file list), `docs/self-review/kan-87-cut-per-command-load-further-self-review.md` (preserved
historical session record, never rewritten), and `commands/myflow-do.md`,
`commands-claude/myflow-do.md`, `rules/myflow-manual-review.mdc` (Task 7's per the brief's
context note).

## Tests run (first pass)

```
scripts/check-references.sh        → "check-references: all referenced sections resolve", exit 0
scripts/test-check-references.sh   → all 29 assertions "ok", "All check-references assertions passed", exit 0
scripts/check-contract-budget.sh   → "BUDGET-OK: 24 contract file(s) within budget", exit 0
scripts/check-vocabulary.sh        → both guards report clean, exit 0
```

---

## Fix round: two Important findings

### Finding 1 — `skills/myflow-do/SKILL.md:8` said "stage both," which is now false

The frontmatter (line 3) already said the handoff prints the run instructions, but line 8 still
read "Implement an OpenSpec change, resolve its run instructions, and stage both for the human
gate at `IN_PROGRESS`" — "stage both" is wrong once the run instructions are only ever printed,
never staged. Fixed to:

> Implement an OpenSpec change, stage the implementation, and print the run instructions in the
> handoff for the human gate at `IN_PROGRESS`: the operator reviews the staged diff and runs the
> apps.

The gate itself is unchanged — still a staged diff to review and apps to run — only the
"stage both" claim is gone, since the run instructions were never a file to stage.

### Finding 2 — the numeric counts, re-derived against the current tree

I re-derived both flagged measurements by writing a script that imports the guard's own functions
(`CLAIM_RE`, `quotation_regions`, `_is_quoted`, `_QUOTE_PAIRS`, `_quote_regions`,
`_has_escaped_delimiter`) via `importlib`, rather than re-implementing the regex/pairing logic
from scratch, so the "lenient" comparison scan shares the real class-wide and escape veto code
and only omits the angle-bracket veto (or, for the first script, omits all three vetoes, per
`plan-provenance.md`'s own described method: "escapes and raw HTML are ignored, a second opener
simply re-arms `pending`, and a leftover `pending` is dropped instead of vetoing the class").
Scripts and full output are in
`/private/tmp/claude-501/-Users-tweety53-Projects-agents/a4f52e3d-0b9f-40eb-8433-e92db3379ce0/scratchpad/measure-vetoes.py`
and `measure-angle-cost.py`.

**`plan-provenance.md:279-286` ("Twelve lines… Eleven are counter-examples"):**

Ran the "all three vetoes removed" comparison against `git ls-files '*.md'` (216 files, current
tree, `docs/manual-test/` already deleted from the index): **12 lines lose an enclosure, 11 of
them are worked demonstrations of a veto (this file, the live delta spec, and archived/preserved
copies of the same shapes), 1 is the pre-existing genuine cost row (the archived
`kan-14-plan-provenance/tasks.md:139` line, unchanged).** That matches the numbers already
written — **no numeric edit was needed.**

I did not take that on faith. I also ran the identical script against the pre-task tree
(`git archive HEAD` extracted to a scratch directory, `docs/manual-test/` still present: 230
files): **17 lines lose an enclosure.** The difference is exactly the 5 lines the now-deleted
`docs/manual-test/kan-20-widen-plan-provenance-guard-scan-scope.md` guide itself contributed
(lines 76, 84, 93, 102, 115 — its own copies of the same demonstrations). So the guide's removal
dropped the true count from 17 to 12, which happens to be the number already on the page: the
page's numbers were correct for the post-deletion tree without editing, because the deletion
exactly cancelled out a pre-existing five-line overcount the page's stated "Twelve" had never
accounted for. I left the numbers as they stand, verified rather than assumed.

**`scripts/check-plan-provenance.py` (the "38 numeric claims… 2 on `<`" comment, `over all 110
markdown files`):**

This is a different measurement — the angle-bracket veto's own cost, i.e. claims that would keep
their exemption under the class-wide and escape vetoes alone, with the angle-bracket veto
specifically removed. I reimplemented that (escape veto verbatim from the module, angle-bracket
veto skipped) and ran it against `git ls-files '*.md'` (216 files, current tree): **37 claims are
exempted, 9 of them sit on a line containing `<`.** The stated "110 markdown files" was already
far from the real count (`git ls-files '*.md' | wc -l` → 216) even before this task — running the
same script against the pre-deletion `git archive HEAD` tree (230 files) gave 42 exempted / 11 on
a `<` line, nowhere near "110/38/2" either. This number had drifted long before `docs/manual-test/`
was touched; the 14 deleted files are a small fraction of the 216-vs-110 gap. I corrected the
comment to the current, reproducible figures (216 files, 37 exempted, 9 on `<`) and named the
method (`git ls-files '*.md' | wc -l`, escape + class-wide vetoes only) so a future reader can
re-run it rather than trust a static number.

**A third, related stale count I found while checking this (not in either flagged location):**
`plan-provenance.md:293-297` ("is two claims across all 110 Markdown files here") is the same
"110" figure, describing yet a different, harder measurement — the cost of the coarse
any-`<` rule against what a *true* CommonMark §6.6 raw-HTML-aware reader would exempt (not just
"any line with a `<`," but only lines where the `<` isn't actually opening a real HTML construct).
Reproducing that precisely needs a genuine CommonMark raw-HTML boundary detector, which — per this
same file's own reasoning elsewhere about not reaching for a tokenizer this repository doesn't
have — I did not build. Rather than leave a number that reads as current, I removed the specific
figure and stated plainly why it isn't re-derived here, keeping the qualitative point (the cost is
real but rare, reword the line if you hit one) intact. This follows the reviewer's stated fallback:
"if the measurement genuinely cannot be reproduced now — say so in the comment itself."

## Tests run (fix round)

```
scripts/check-references.sh         → "check-references: all referenced sections resolve", exit 0
scripts/test-check-references.sh    → all assertions "ok", exit 0
python3 scripts/check-plan-provenance.py → "check-plan-provenance: 3 file(s) scanned, all provenance stated", exit 0
scripts/check-contract-budget.sh    → "BUDGET-OK: 24 contract file(s) within budget", exit 0
scripts/test-check-plan-provenance.sh → all assertions "ok", "All check-plan-provenance assertions passed", exit 0
scripts/check-vocabulary.sh         → both guards report clean, exit 0
grep -rln "manual test guide\|manual-test" skills/ scripts/ docs/ commands/ commands-claude/ rules/ | grep -v 'docs/superpowers/'
  → only scripts/check-vocabulary.sh (allowlist), scripts/test-check-unfinished-work.sh (Tasks
    1-5's own test), docs/self-review/kan-87-...-self-review.md (preserved history), and Task 7's
    three files (commands/myflow-do.md, commands-claude/myflow-do.md, rules/myflow-manual-review.mdc)
```

## Files changed (fix round, in addition to the first pass)

- `skills/myflow-do/SKILL.md` (line 8)
- `skills/myflow-contracts/plan-provenance.md` (lines 293-297, the "two claims" paragraph;
  279-286 verified unchanged)
- `scripts/check-plan-provenance.py` (the "cost was measured" comment, ~lines 586-596)

No commits made in either pass. All changes remain uncommitted in the worktree.
