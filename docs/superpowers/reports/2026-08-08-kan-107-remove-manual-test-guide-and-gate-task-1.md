## Task 1 report: Drop the guide signals from the unfinished-work guard

**Status:** DONE

### Files changed

- `scripts/check-unfinished-work.sh`
- `scripts/test-check-unfinished-work.sh`

### What was done

Followed RED-GREEN-REFACTOR: the harness change landed first, was run and confirmed to fail for
the stated reason (guard still required the guide), then the guard was edited, then the harness
was run again and passed in full, then the two surviving signals were broken and restored by hand
to confirm they still fire.

**Harness (`scripts/test-check-unfinished-work.sh`):**
- `new_fixture()` no longer creates `$WT/docs/manual-test` or `demo.md`; its comment was trimmed to
  describe only the plan and panel fixtures it still builds.
- Deleted every case that planted, mutated, chmodded or removed `$WT/docs/manual-test/demo.md`:
  the old "Signal one" unticked-guide-box case, both "Signal four" Known-incomplete cases, the
  "missing guide" half of the missing-file group (kept the "missing panel record" half), both
  missing/empty Known-incomplete-section cases, the combined-signals case (it wrote to the guide),
  and the whole 8b/8c/8c-i/8c-ii/8c-iii/8c-iv fence-tracking suite plus the 8d unreadable-guide
  case.
- Added the one new case the brief asked for, in the freed slot 2: `rm -rf "$WT/docs/manual-test"`
  then `run_guard "$WT" demo` then asserts `CLEAR:`.
- Left `8e`/`8e-i` (allowlist) and `8f` (traversal into a planted `docs/manual-test` tree under a
  *different* sandbox dir, not `$WT`) untouched — they don't touch `$WT/docs/manual-test/demo.md`
  and still exercise real behavior (the allowlist and traversal containment are unaffected by this
  change).
- Fixed one now-stale header line: "collapsed all four signals" -> "collapsed both signals".

**Guard (`scripts/check-unfinished-work.sh`):**
- Deleted the `GUIDE=` assignment.
- Deleted the whole "Signals one and four" block: the `[ ! -f "$GUIDE" ]` branch, `BOXES`, and the
  `SECTION_STATE` awk program with its three `KNOWN_*` branches.
- Deleted the fence-tracking rationale block in the header comment ("WHERE FENCE TRACKING IS PAID
  FOR" / "WHAT THE FENCE RULE ACTUALLY IS") — it existed solely to justify the `## Known incomplete`
  scanner, now gone.
- Trimmed the "A MISSING FILE COUNTS AS OUTSTANDING" header paragraph and the "All four signals"
  sentence down to the two surviving signals.
- Kept `count_unticked()` and `unreadable()` — both are still called by the plan signal (and
  `unreadable()` by the panel signal too) — and reworded `count_unticked()`'s comment from "the
  guide and the plans are the same knowledge" to "the primary plan and every fix sub-change's plan
  are the same knowledge".
- Renumbered every surviving in-code comment: "Signal two — the plan" -> "Signal one — the plan";
  "Signal three — findings" -> "Signal two — findings", plus the two prose cross-references to
  "signal three" inside that section's body ("Signal three no longer compares strings" and
  "signal three has no notion of a block").
- Fixed a stray "no guide at all" in the CONTAINMENT comment (traversal example) to "no plan of its
  own", since the guide path is gone but the plan path (`$CHANGES/$NAME/tasks.md`) is still built
  from the same PR-controlled `$NAME`.
- Updated the final `CLEAR:` message from "every checklist box is ticked, every plan item is
  checked, no finding is open, and nothing is recorded as incomplete" to "every plan item is
  checked and no finding is open".
- Exit-code contract unchanged: `0` for a verdict (`CLEAR`/`OUTSTANDING`), `2` when it cannot
  determine, no verdict line on `2`.

### Plan-provenance note (Step 1 / Step 5 `unverified:` hypotheses)

Read `scripts/test-check-unfinished-work.sh` before writing anything. Confirmed:
- `run_guard "$worktree" "$name"` sets `OUT`/`ERR`/`RC` — matches the brief's hypothesis.
- `$WT` is the fixture variable set by `new_fixture()` — matches.
- The change name used throughout is `"demo"` — matches.
- There is **no `assert_clear` helper**. The brief's Step 1 snippet's `assert_clear "..."` does not
  exist in this file; the harness uses `assert_verdict "CLEAR:" "<label>"` for that purpose (see
  `assert_verdict()`'s definition and every existing CLEAR-asserting case, e.g. case 1). The new
  case was written with `assert_verdict "CLEAR:" "an absent docs/manual-test/ is not a signal"`
  instead of the hypothesized `assert_clear`.
- Step 5's snippet was accurate as a description of what should happen; it was not literally
  transcribed into a `bash unverified:` fenced block executed unverified — it was run by hand
  against real fixtures (see below) instead of pasted into the test file, since the brief only asked
  to "prove the surviving signals still fire," not to add permanent cases for it.

### Test summary

- `scripts/test-check-unfinished-work.sh`: RED before the guard edit (10 failing cases, all failing
  with the same reason — `OUTSTANDING: ... no manual test guide at .../docs/manual-test/demo.md` —
  proving the harness was exercising the intended contract). GREEN after: `check-unfinished-work:
  all cases pass`.
- Manual proof of Step 5 (both surviving signals, done twice — break then restore):
  - Baseline (clean plan + clean panel) -> `CLEAR: ... every plan item is checked and no finding is
    open`.
  - Appended `- [ ] a task` to `tasks.md` -> `OUTSTANDING: ... 1 unchecked plan item(s)`. Restored,
    back to CLEAR.
  - Changed the panel's `finding-status: F1 fixed` to `finding-status: F1 open` -> `OUTSTANDING: ...
    1 open finding(s) in the review panel record`.
- Lint: `scripts/check-vocabulary.sh` (clean), `scripts/check-references.sh` (all referenced
  sections resolve), `scripts/check-plan-provenance.sh` (3 files scanned, all provenance stated),
  `scripts/check-contract-budget.sh` (24 contract files within budget — this task's two files
  aren't budget-tracked and both shrank, so no ratchet row needed raising). `bash -n` syntax check
  clean on both files. `shellcheck` is not installed in this environment, so it was skipped.

### Concerns

None. `docs/manual-test/demo.md` itself was left untouched on disk, as instructed (deletion is
Task 6's job) — the guard and its test harness simply no longer look at it.

## Fix round 1 (review findings)

Review of the task came back with the spec satisfied and three findings: one Important, two
Minor. All three fixed in this round; no new suppressions, no guard weakening, no commits, and
nothing under `docs/manual-test/` was touched.

### Important — stale "Signal N" cross-references in `scripts/test-check-unfinished-work.sh`

The guard's own comments were renumbered in the original pass, but the test file's mirror
comments still cited the old four-signal numbering in four places. Fixed all four, plus confirmed
by re-grepping the file that no other signal-number reference remained:

- Line 29 (header prose, about the findings signal): "Signal three was then redesigned so that
  none of those questions is asked at all" -> "Signal two was then redesigned so that none of
  those questions is asked at all".
- Line 200: `# 3. Signal two — the plan.` -> `# 3. Signal one — the plan.`
- Line 270: `# 4. Signal three — findings whose recorded status is not closed.` -> `# 4. Signal
  two — findings whose recorded status is not closed.`
- Line 272: "READ THE GUARD'S HEADER FIRST. Signal three does not parse the human findings..." ->
  "...Signal two does not parse the human findings...".

A follow-up grep for `signal one|signal two|signal three|signal four` (case-insensitive on the
word "Signal") across the whole file turned up only these same four sites, now corrected — no
other stale numbering was found. (The unrelated numbered list inside the findings-signal header
comment — "1. NO DECLARED TOTAL", "2. A TOTAL DECLARED TWICE", etc. — enumerates sub-cases of that
one signal, not the four top-level signals, and was correctly left alone.)

### Minor — stale guide prose in case 8e

Case 8e's rationale comment (just above the bad-name loop) still said: "Without it
`../../../planted/clear` reads a guide outside the worktree entirely and reports CLEAR for a
change that has none." The guard no longer builds any guide path — `GUIDE` was deleted in the
first pass — so the sentence was reworded to describe what the guard now protects: "Without it
`../../../planted/clear` reads a plan outside the worktree entirely and reports CLEAR for a
change that has none." The assertion block below it (loop over `bad_name` values, checking `RC
-eq 2`, empty `OUT`, and the stderr message) was left exactly as it was.

### Minor — vestigial guide fixture in case 8f

Case 8f built a planted `$PLANTED/docs/manual-test/clear.md` file that nothing in the case reads:
the guard rejects the traversal-shaped name (`$PLANTED/docs/manual-test/clear`, which contains
`/`) via the NAME allowlist before it opens any file, and the case's only assertion is that `OUT`
is empty. Removed the vestigial planting:

- Dropped `printf '# ok\n\n## Known incomplete\n\nNone.\n' > "$PLANTED/docs/manual-test/clear.md"`
  entirely.
- Dropped `"$PLANTED/docs/manual-test"` from the `mkdir -p` line, keeping
  `"$PLANTED/openspec/changes/clear"` and `"$PLANTED/.superpowers/sdd"` (those still stand in for
  a "finished change" outside the worktree, which is what the case's own comment describes).

The case itself, its `run_guard "$WT" "$PLANTED/docs/manual-test/clear"` call (the traversal-shaped
name being tested, not a path the guard actually reads), and its assertion were left intact, as
instructed.

### Covering test command and output

```
$ bash -n scripts/test-check-unfinished-work.sh && echo "syntax ok"
syntax ok
$ scripts/test-check-unfinished-work.sh
```

Full run: every case reports `ok:`, ending with:

```
check-unfinished-work: all cases pass
```

No `FAIL:` lines anywhere in the output. Also re-ran the lint set named in `.myflow/project.md`
for the two changed files' neighborhood — `scripts/check-vocabulary.sh` ("Stage-vocabulary guard:
clean", "Panel-vocabulary guard: clean"), `scripts/check-references.sh` ("check-references: all
referenced sections resolve"), `scripts/check-plan-provenance.sh` ("check-plan-provenance: 3
file(s) scanned, all provenance stated"), and `scripts/check-contract-budget.sh` ("BUDGET-OK: 24
contract file(s) within budget") — all clean, no regressions from this round.
