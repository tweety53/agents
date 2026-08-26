# Tasks

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.

One guard and its harness. Task 1 is the collapse itself — guard code, guard header and the one
harness assertion that moves — kept as a single commit because a commit that deletes the record
protocol while leaving the header documenting it is incoherent on its own. Task 2 re-earns the
mutation proof KAN-211 credits the harness with and records the branch-to-case matrix in the
harness header, so a later reader can check the proof rather than take it on trust.

- [x] 1. Collapse the awk→bash boundary in `check-self-review-report.sh`

  Delete `scan_report()` (currently the `awk` heredoc around lines 286-320) and the
  `while IFS=$'\037' read -r kind label lineno rest` loop that consumes it (lines 380-470). In their
  place, read each report directly inside the existing per-file `for f in "${FILES[@]:-}"` loop:
  `lineno=0; cur=""` then `while IFS= read -r line || [[ -n "$line" ]]; do` … `done < "$f"`, with
  `line="${line%$'\r'}"` as the first statement in the body so CRLF reports still read as LF ones
  (harness case 18). Inside the loop, reproduce exactly the four branches the record protocol
  encoded, in this order: a `^##[[:space:]]` line sets `cur` to whichever of the five
  `ANGLE_LABELS` appears backtick-quoted in it (or `""` when none does) and runs the existing
  `HEADING` classification body; any other `^#+[[:space:]]` line sets `cur=""`; a line with `cur`
  non-empty runs the existing `BODY` classification body against `$line` in place of `$rest`; a line
  with `cur` empty that matches `$FINDING_SHAPE_RE` emits the existing orphan violation. Delete
  `FINDING_SHAPE_RE_AWK` and its comment outright — with no `-v` relay there is no second spelling
  to derive. Delete the `CTRLBYTE` branch and its violation entirely, per `retire-ctrlbyte-rule`:
  `0x1F` is now ordinary report content.

  Keep untouched: `ANGLE_LABELS`, `FINDING_SHAPE_RE`, `FINDING_LINE_RE`, `DISPOSITION_FILED_RE`,
  `ISSUE_KEY_RE` and the `[[ =~ ]]` quote-removal comment on the first of them; `DECLARED_REPORTS`,
  `declare_pre_rule`, `is_declared_basename` and both `$# -eq 0` gates; the `find`-to-temp-file
  enumeration and its exit-status check; the per-angle `HEADING_FOUND`/`HEADING_LINE`/`HAS_NONE`/
  `FINDING_COUNT` arrays and `LAST_IDX`; every `coverage_*` call; the post-loop missing-section,
  neither-marker-nor-finding and both-marker-and-finding checks; every message string; and all three
  exit codes.

  Rewrite the header in the same commit. Delete the `US`-versus-tab protocol paragraph, the awk `-v`
  backslash-doubling paragraph, the `CTRLBYTE` record description and the `scan_report` record-shape
  block. Replace them with a short note recording that an `awk`-to-`bash` record boundary existed
  here, that it produced KAN-200's G2, G3 and H2, and that KAN-211 collapsed it rather than patching
  a fourth manifestation — so a future reader does not reintroduce one. Update the `ADOPTED, NOT
  INVENTED` paragraph, which currently says the guard uses no `grep` because extraction is done "in
  one `awk` pass per file, then classified in bash": it now uses neither `grep` nor `awk`, and the
  `find` call is the only remaining exit status to check.

  In `scripts/test-check-self-review-report.sh`, change case 21's second assertion so it matches on
  the section-level violation instead of the retired one — the same pattern shape case 20 already
  uses, naming the offending angle and then the phrase "neither a finding line nor the none-marker"
  — and rewrite its comment block to say what the case now proves: the `0x1F` byte reaches the
  classifier byte for
  byte, so a none-marker line carrying a trailing one no longer equals the none-marker and its
  section is flagged. Leave the fixture, the `grep -q $'\037'` setup assertion and the exit-1
  assertion exactly as they are. Do not touch cases 1-20.

  **Add one further assertion to case 21: the phrase `unsupported control byte` does not appear in
  the output at all.** The retargeted assertion above is a correct statement about the collapsed
  guard, but it is not a regression detector: the unmodified guard reports *both* violations for
  this fixture — the control-byte one, and the section-level one, because `CTRLBYTE` consumed the
  none-marker line — so the retargeted assertion alone passes in both states and a revert of this
  task would go uncaught. Asserting the retired violation's absence is what makes the revert loud,
  and it is the only assertion in the harness that can do so.

  Verify: `scripts/test-check-self-review-report.sh` prints `all cases passed`;
  `scripts/check-self-review-report.sh` exits 0 with its `SELF-REVIEW-REPORT-OK` verdict over the
  real corpus; and no **non-comment** line of `scripts/check-self-review-report.sh` matches
  `awk`, `037` or `FINDING_SHAPE_RE_AWK` — check with
  `grep -n 'awk\|037\|FINDING_SHAPE_RE_AWK' scripts/check-self-review-report.sh | grep -v '^[0-9]*:[[:space:]]*#'`,
  which must return nothing. The bare grep is deliberately **not** the test: this same task requires
  the header to record that an awk-to-bash boundary existed here, so header prose naming `awk` is
  the instruction being followed, not a leftover. Then run the project's `## lint` list and confirm
  it is clean. Measure the
  guard's runtime over the real corpus and report it against the design's recorded baseline; if it
  has grown materially, report the measurement rather than absorbing it.

**Build:** green
**Files:** `scripts/check-self-review-report.sh`, `scripts/test-check-self-review-report.sh`
**Tests:** no new case — `scripts/test-check-self-review-report.sh` case 21's message assertion is
retargeted and one assertion is added to that same case; cases 1-20 are unchanged and are the
regression surface for the rest of this task.
**Regression:** reverting this task restores `scan_report`, the `US` record protocol and the
`FINDING_SHAPE_RE_AWK` relay, reopening the G2/G3/H2 bug class KAN-211 exists to close. Case 21's
added assertion — that `unsupported control byte` is absent from the output — fails immediately on
such a revert, because the restored guard emits that violation again. The retargeted assertion
beside it does **not** catch the revert on its own: the pre-change guard reports the section-level
violation too, since `CTRLBYTE` consumed the none-marker line, so it passes in both states. That was
measured on the unmodified tree, not assumed, and it is why the added assertion exists.
**Baseline:** before=43 after=44 assertions in `scripts/test-check-self-review-report.sh`; the guard
exits 0 over the real corpus in both states.
<!-- measured: scripts/test-check-self-review-report.sh | grep -c '^ok:' @ branch main, before any edit -->
**Commit:** `refactor(scripts): collapse check-self-review-report's awk to bash boundary`

- [x] 2. Re-earn the mutation proof and record the branch-to-case matrix

  With the classifier now in one language, prove each of its branches is actually covered. Working
  from a clean tree, revert one classification branch at a time in
  `scripts/check-self-review-report.sh` — the `##`-heading label match, the other-heading reset, the
  none-marker comparison, the `FINDING_LINE_RE` acceptance, the label-mismatch check, the
  `declined`/`filed:` disposition split, the issue-key shape check, the loose-shape malformed
  branch, the orphan branch, the duplicate-section check, the out-of-order check, the
  missing-section check, the neither-marker-nor-finding check and the both-marker-and-finding check
  — running `scripts/test-check-self-review-report.sh` after each and recording which case fails.
  Restore the branch before mutating the next; never commit a mutated tree.

  A branch whose reversion leaves every case passing is an uncovered branch and is reported as a
  finding of this task, not silently accepted: either it is dead code to delete or it needs a case,
  and which one it is is a decision for the operator rather than something to resolve inside this
  task.

  Record the result as a comment block in `scripts/test-check-self-review-report.sh`'s header: one
  line per branch naming the case number that catches its reversion. This is what makes the proof
  checkable later instead of a claim in a commit message.

  **Also carry one leftover from task 1**, which could not take it: case 12's comment
  (`scripts/test-check-self-review-report.sh:336`) says a line "must surface as an **ORPHAN** (no
  section open)". `ORPHAN` was a literal record-kind token the deleted `awk` printed; no such token
  exists now, only the descriptive concept. Lowercase it, exactly as task 1 lowercased the two
  equivalent mentions in the guard. Task 1 forbade touching cases 1-20 and correctly left it; this
  task already owns the file. Leave case 21's own comment alone — its backticked, past-tense
  `CTRLBYTE` describes the pre-change guard it compares against, and is correct as written.

  Verify: `scripts/test-check-self-review-report.sh` prints `all cases passed` on the restored,
  unmutated tree, and the recorded matrix names a case for every branch listed above.

**Build:** green
**Files:** `scripts/test-check-self-review-report.sh`
**Tests:** no new case — this task adds the mutation matrix to the harness header; the assertions
themselves are unchanged from task 1.
**Regression:** reverting this task removes the recorded matrix, returning the harness to a state
where its mutation coverage is asserted in prose elsewhere but cannot be checked against the code —
which is how KAN-200's own seam defects survived four review rounds.
**Baseline:** before=44 after=44 assertions — task 1 leaves the harness at 44; a header comment adds
no assertion.
<!-- measured: scripts/test-check-self-review-report.sh | grep -c '^ok:' @ branch main, before any edit -->
**Commit:** `test(scripts): record check-self-review-report's mutation matrix`

- [x] 3. Cover the two branches the mutation proof found uncovered

  Task 2 measured two live, reachable classification arms that no fixture reaches: deleting either
  leaves every assertion green. Add one case per branch, appended after case 21 as cases 22 and 23,
  per `design.md`'s `cover-the-two-uncovered-branches`. Follow the file's existing case shape
  exactly — the banner comment, `new_fixture`, a heredoc or `printf` fixture, `run_guard`, then an
  exit-status assertion and a message assertion, each with its own `pass`/`fail`.

  **Case 22 — a disposition that is neither `filed: <KEY>` nor `declined`.** Build a compliant
  report whose `myflow-cost` finding line ends `— maybe later`. The line is a well-formed finding
  line, so `FINDING_COUNT` increments and the section is not otherwise flagged; deleting the
  disposition arm therefore drops the guard to exit 0. Assert exit 1 and the message ``finding line
  disposition is neither `filed: <KEY>` nor `declined` ``.

  **Case 23 — a finding-shaped line that is malformed.** Build a report whose `myflow-fix` section
  carries only a line with **two** spaces after the leading `-` (`-  **[myflow-fix]** … — declined`)
  and no none-marker. Assert exit 1 **and** the message `finding-shaped line is malformed`. **The
  exit-status assertion alone does not detect this branch's removal** — with the loose-shape arm
  deleted the line is recognized as no finding at all, so the post-loop neither-marker-nor-finding
  check fires on the same fixture and holds the exit code at 1. The message assertion is what makes
  the removal loud; say so in the case's comment, as case 21 already does for its own pair.

  This is the branch KAN-200 added as its G2 fix, so that a malformed finding-shaped line would stop
  being silently dropped. State that in the case's comment: it records why the case exists and what
  regression it guards.

  Prove each new case actually detects its branch's removal before trusting it, exactly as task 1
  proved case 21's absence assertion: delete the branch, run the harness, confirm **that** case
  fails; restore the branch, confirm the harness is green again. Never commit a mutated tree.

  Then extend the mutation matrix task 2 wrote into the harness header: the two rows currently
  reading `none — uncovered` name their new case numbers, and the paragraph describing the
  masking trap is updated to point at case 23 as the case that handles it.

  Verify: `scripts/test-check-self-review-report.sh` prints `all cases passed`;
  `scripts/check-self-review-report.sh` exits 0 with its `SELF-REVIEW-REPORT-OK` verdict over the
  real corpus; `git diff aa7dfe8 HEAD -- scripts/check-self-review-report.sh` is empty. Then run the
  project's `## lint` list and confirm it is clean.

**Build:** green
**Files:** `scripts/test-check-self-review-report.sh`
**Tests:** two new cases — case 22 (disposition neither `filed:` nor `declined`) and case 23
(finding-shaped line malformed), each asserting exit status and its own violation message.
**Regression:** reverting this task returns both branches to zero coverage, so deleting either one
again passes the whole harness — including the arm KAN-200 added as its G2 fix, whose entire purpose
is that a malformed finding-shaped line is not silently dropped.
**Baseline:** before=44 after=48 assertions in `scripts/test-check-self-review-report.sh` — two
cases contributing two assertions each; the guard still exits 0 over the real corpus.
<!-- predicted: scripts/test-check-self-review-report.sh | grep -c '^ok:' once this task's two cases land -->
**Commit:** `test(scripts): cover check-self-review-report's two untested branches`
