# SDD ledger — plan: openspec/changes/kan-17-finish-gate-jira-and-commit-hygiene/tasks.md

Change: kan-17-finish-gate-jira-and-commit-hygiene
Worktree: /Users/tweety53/Projects/agents-worktrees/openspec-kan-17-finish-gate-jira-and-commit-hygiene
Merge base: d0c1121db2f12889d243844db50cebbaf4fffb8e
Git boundary: NO COMMITS this run (pipeline.md — /myflow-do stages only; no prUrl recorded).
Model policy: implementers on Opus (myflow override of SDD's cheapest-tier guidance); every panel slot on Sonnet.

Pre-flight plan scan: no task contradicts another or the Global Constraints. The plan's
"No per-task commits" constraint deliberately overrides the writing-plans per-task commit
step and says so in place; that is a recorded override, not a conflict.

## Progress

Task 1: implementer DONE_WITH_CONCERNS (uncommitted, model: opus). 28/28 harness cases pass;
  RED watched first (25/28 failing pre-guard); 8 single-edit mutations all caught.
  Implementer corrected a real defect in the plan's `unverified:` sketch: the unanchored
  `- \[ \]` pattern also matched checklists quoted inside fenced examples (44 unanchored vs
  42 anchored hits on this change's own tasks.md). Pattern now anchored to line start.
  Two widenings beyond the brief, both sent to the reviewer for explicit assessment:
    (a) an EMPTY `## Known incomplete` section counts as OUTSTANDING (spec's letter names
        only "non-empty" and "missing");
    (b) the guard reads both `<name>-fix-*/` and `<name>/<name>-fix-*/` layouts, because
        no contract fixes which one a fix run writes and guessing wrong yields a false CLEAR.
Task 1: reviewer dispatched (model: sonnet), diff .superpowers/sdd/task-1.diff (400 insertions).

Controller finding, plan defect (not the implementer's): skills/myflow-do/SKILL.md:191
  ("Union all Critical/Important findings ... give one fix subagent") also needs widening
  once minors block, and Task 6's brief does not name it. Task 6's step 4 verification would
  have been misleading as written. Fold into Task 6's brief and into tasks.md before dispatch.
Task 1: reviewer verdicts — spec ✅, task quality Approved (model: sonnet).
  Reviewer independently reproduced one row of the implementer's mutation table (dropped the
  nested-fix glob -> exactly one failing case), corroborating the TDD record rather than taking
  it on trust. Both widenings explicitly endorsed.
Task 1: minor (deferred): check-unfinished-work.sh:114-116 interpolates $NAME unescaped into a
  glob; a change name containing * ? or [ would misbehave. Pipeline-controlled slug, low risk.
Task 1: minor (deferred): test-check-unfinished-work.sh case 8 fires three signals but asserts
  only two — under-specified, not wrong.
Task 1: complete (uncommitted, review clean, model: opus; reviewer sonnet).

Controller edit between tasks (not attributable to any implementer): tasks.md Task 6 amended —
  step 2 now also widens skills/myflow-do/SKILL.md:191 ("Union all Critical/Important findings"),
  and step 4's expectation corrected from "matches only in reviewer prompts" to "no match in
  SKILL.md". A bar blocking on every severity while the fix wave collected only two of them would
  have left the run unable to clear its own gate. Task 6's brief regenerated from the amended plan.
Task 2: implementer DONE_WITH_CONCERNS (uncommitted, model: opus). 33/33 harness cases pass
  (RED first: 22 failing pre-guard); Task 1's 28/28 still pass; three lint guards exit 0.
  Corrected THREE defects in the plan's `unverified:` sketch:
    - the awk read the worktree path as $2; `worktree list --porcelain` emits paths raw, so a
      path containing a space was truncated. Now substr($0,10), pinned by a mutant-killing case.
    - `| awk ... || true` turned an UNREADABLE repository into COMPLETE — a reassuring verdict
      produced by a failure. Now an explicit capture and exit 2.
    - the sketch's case list claimed "one per registry row" but covered 3 of 5: no worktree case
      and no remote-tracking-ref case (the row this change adds). Both added.
  Mutation-tested; one mutant survived round 1 (loose local-branch matching reported the local
  row from a surviving remote ref alone), so assert_no_reason plus two cases were added.
Task 2: reviewer dispatched (model: sonnet), diff .superpowers/sdd/task-2.diff (448 insertions).

Controller edit between tasks: tasks.md Task 10 gained a new Step 2 — name check-unfinished-work.sh
  and check-cleanup-complete.sh in .myflow/project.md's `## lint` rationale, which explains why
  in-flight-only guards are not lint steps. Task 2 added them to `## test` but was explicitly told
  not to touch `## lint`, so the paragraph now under-describes its own exclusion list. Old Step 3
  renumbered to Step 4; Task 10's brief regenerated.
Task 2: reviewer verdicts — spec ✅ (all five registry rows), task quality Approved (model: sonnet).
  Reviewer re-ran both harnesses (33/33 and 28/28) and independently reproduced the git
  worktree-listing finding: chmod 000, a corrupted gitdir pointer, a missing HEAD and a broken
  symlink all leave `worktree list --porcelain` exiting 0. All three flagged judgment calls agreed.
Task 2: minor (deferred): check-cleanup-complete.sh:~97-100, the worktree-listing failure branch is
  defensive-only and unreachable in practice — harmless.
Task 2: complete (uncommitted, review clean, model: opus; reviewer sonnet).

Controller edit between tasks: tasks.md Task 9 gained a new Step 3 — fix pipeline.md:340's
  fallback worktree scan, which reads the path as awk's $2 and truncates any path containing a
  space (verified: `worktree /tmp/my worktree/x` yields `/tmp/my`). Run 2 would then --force-remove
  a path that is not the worktree. Surfaced by Task 2's reviewer as out-of-scope-but-latent;
  check-cleanup-complete.sh already parses that stream correctly, so the contract snippet was the
  one disagreeing with its own verifier. Old steps 3-5 renumbered to 4-6; Task 9's brief regenerated.
Task 3: implementer DONE_WITH_CONCERNS (uncommitted, model: opus). Requirement checks went
  15 pass / 19 fail -> 34 pass / 0 fail; all three lint guards and all seven `## test` harnesses
  exit 0; a region diff against HEAD proves the Never-blocking and description-sync sections are
  byte-for-byte unchanged, as required.
  Found and fixed a defect outside the brief: this file's only outbound cross-reference
  `(see **State file**, state-file.md)` had an unbackticked path and was NOT being checked —
  proved by mutation (guard exited 0 with the section deliberately renamed), rewritten to a
  checked shape, re-mutated, guard now exits 1.
  TRAP TO CARRY FORWARD (applies to Tasks 4, 8, 9, 10): check-references.sh scans ONE PHYSICAL
  LINE at a time. Reflowing a paragraph so the bold token and the path land on different lines
  silently un-checks the reference. The implementer hit this on its first attempt.
Task 3: reviewer dispatched (model: sonnet), diff .superpowers/sdd/task-3.diff (34 insertions,
  5 deletions).

Controller-owned check, deferred to the final verification (not a plan amendment): the implementer
  reports the four interface heading names have no real guard behind them, because every referring
  line also carries **Jira integration**, which matches the file title, and check-references.sh
  needs only one associated token to match. Task 3's reviewer is verifying that claim. If it holds,
  the controller greps the four headings directly after Task 9 rather than trusting the guard.
Task 3: reviewer verdicts — spec ✅ (all seven requirements mapped to lines), task quality Approved
  with minors (model: sonnet). Reviewer independently mutation-tested items 3 and 5 and confirmed
  both; it also corrected the implementer's diagnosis of the unchecked reference — backticking the
  path alone is sufficient, the comma connector was not a second cause.
Task 3: CONFIRMED GAP (controller-owned, verify after Task 9): the four interface heading names have
  no guard behind them. The reviewer renamed `### Description sync` to a typo, left the referring
  line in myflow-do/SKILL.md untouched, and check-references.sh still exited 0 — the neighbouring
  **Jira integration** token matches the file title and satisfies the one-associated-token rule.
  A Task 4/8/9 misspelling of **Transitions**, **Unrecognised statuses** or **Labels on issues the
  pipeline creates** would therefore pass silently. Controller greps all four directly after Task 9.
Task 3: minor (deferred): task-3-report.md:13 says "nothing was staged"; the tree IS staged, which
  is the correct state. Report phrasing only, no file defect.
Task 3: complete (uncommitted, review clean, model: opus; reviewer sonnet).
Task 4: implementer DONE_WITH_CONCERNS (uncommitted, model: opus). All three guards clean,
  test-check-references.sh 31/31, brief Step 4 grep clears. Mutation test on the file title made
  the guard fail on both new reference lines and pass again after restore.
Task 4: MAJOR FINDING, repository-wide — check-references.sh refuses to associate a bold token with
  a path when ANOTHER `**bold**` sits in the gap. So the pervasive shape
  `**Subsection** in **Jira integration** (`path`)` checks ONLY the file title, and every subsection
  name referenced that way is an unguarded interface. This is the mechanism behind the gap Task 3's
  reviewer proved by mutation. The checked shape is `**Subsection** (`path`)` — plain-text file
  names are fine, only a second bold token breaks it. skills/myflow-start/SKILL.md:71 already uses
  the correct shape for **Effort**.
Task 4: fix round 1/5 dispatched (resumed the same implementer): add the spec's "at every effort
  level" clause (scenario "A low effort level still asks" was unsatisfied by any line), and rewrite
  its two new references into the checked shape with a real mutation test on each.
Task 4: minor (deferred): jira-integration.md's Description sync requires the pre-edit description
  be echoed verbatim into the handoff, and /myflow-start's handoff block has no slot for it.
  Out of scope for this task; surface to the final panel.
Task 4: fix round 1/5 complete (2 addressed, 0 open). Effort clause added to the questions
  guardrail; both new references rewritten to the checked shape and EACH seen to fail under a
  heading rename, then restored byte-identical. Counter-check: renaming the file title no longer
  implicates either line (it did before the fix), which is the proof the subsection name is now
  the load-bearing token.
  Shape landed on, for the sweep: `**Subsection** in <plain-text title> (`path`)`, or bare
  `**Subsection** (`path`)`. Gap budget is <= 60 chars AND <= 3 words: " in Jira integration " is
  exactly three words and fits; "in the Jira integration contract" does not and silently restores
  the bug.
Task 4: reviewer dispatched (model: sonnet), diff .superpowers/sdd/task-4.diff.

Controller edit between tasks: tasks.md Task 10 gained Step 3 — sweep every subsection
  cross-reference into the checked shape, with the grep that finds them, the two acceptable shapes,
  the three-word budget, and a required mutation test per rewrite. Two lines match repo-wide today,
  one of them skills/myflow-start/SKILL.md:37 which Task 4 deliberately left. Old steps 3-4
  renumbered to 4-5; Task 10's brief regenerated.
Task 4: reviewer verdicts — spec ✅ (every requirement and scenario mapped to a line), task quality
  Approved with two minors (model: sonnet). Reviewer read is_associated() directly and confirmed
  BOTH limits: `length(gap) > 60` returns 0, and the word count returns `c <= 3`, plus a gap
  containing `**` is rejected outright. It reproduced both mutations and the counter-check exactly,
  restoring each heading and verifying md5 against the original.
Task 4: minor (deferred): SKILL.md:58-59 locally echoes the non-blocking policy already stated in
  the Guardrails bullet at :228 — brief's own prose, matches the file's existing pattern, but the
  two could drift.
Task 4: minor (deferred): "Jira integration" is now styled bold at :37,:43 and plain at :57,:190;
  a future editor "normalising" it would silently revert the reference check. Task 10 Step 4
  resolves this repo-wide.
Task 4: complete (uncommitted, review clean, model: opus; reviewer sonnet; 1 fix round).

Controller edit between tasks: tasks.md Task 10 gained Step 3 — a handoff slot for the echoed
  pre-edit description in both myflow-start and myflow-do. Task 4's reviewer found this is the ONLY
  requirement in the change's delta specs that no task implements: jira-integration.md states the
  echo duty, but neither handoff block has anywhere to put it, so an implementer following the
  templates literally would sync a description and echo nothing. Task 10's Files list widened
  accordingly; later steps renumbered to 4-6; brief regenerated.
Task 5: implementer DONE_WITH_CONCERNS (uncommitted, model: opus). 11 per-scenario greps went
  FAIL->PASS (baseline 10 FAIL, final 0 FAIL); three lint guards exit 0; the Step 4 grep leaves
  exactly one match, the new staging pathspec itself, which is a staging path not a diff filter.
  Re-ran the plan's `verified:` pathspec by hand and confirmed it, then ran the whole two-commit
  sequence end to end and confirmed the split. The prUrl commit exception now splits too.
Task 5: reviewer dispatched (model: sonnet), diff .superpowers/sdd/task-5.diff.

Controller edits between tasks, both from Task 5's forward-looking concerns:
  - tasks.md Task 8 gained Step 5 — pipeline.md's ordering-asymmetry paragraph still says hoisting
    the preserve call would "create AND STAGE docs/superpowers/ files". Since Task 5 excluded that
    path from staging, the staging half is impossible, so the paragraph now asserts something FALSE
    rather than merely stale. Old Step 5 renumbered to 6.
  - tasks.md Task 8 Step 3's own prose carried the BROKEN reference shape
    `**Transitions** in **Jira integration** (`path`)`. Rewritten to `**Transitions** (`path`)`
    before dispatch, with the mutation-test instruction attached — the plan would otherwise have
    instructed an implementer to write an unchecked reference.
  - tasks.md Task 10 Step 1 widened: AGENTS.md and CLAUDE.md each carry their own copy of the
    command table and both still summarise /myflow-do as ending in `git add -A`, which Task 5 made
    false. Both are also this repo's `## standards` entries, so a stale table there is what the
    principles reviewer reads as the project's stated rules. Files list updated; brief regenerated.
Task 5: noted, no action: final-review.diff is still built from `git diff <merge-base>` (staged and
  unstaged), so the review panel still sees openspec/. Deliberate and correct — the panel should
  read the plan. It is now the only diff in the pipeline containing planning artifacts.
Task 5: reviewer verdicts — spec ✅ within the task's file scope, task quality: 1 Critical,
  2 Important, 1 Minor (model: sonnet).
  CRITICAL: the "never staged" guarantee is not self-enforcing. `git add` pathspecs control what is
  ADDED and never retract prior staging, and myflow-do/SKILL.md:80-81 (section 4, untouched by this
  task) tells every implementer dispatch "You **may** `git add`" with no exclusion. One bare
  `git add -A` upstream therefore leaves planning paths staged and makes SKILL.md:228 and
  pipeline.md:150 false. Reviewer reproduced it in a scratch repo.
  IMPORTANT x2, both report-accuracy rather than file defects: the report miscounted which files
  literally say `git add -A` (README.md:349 says plain `git add`); and its justification for
  reformatting the **Finish contract** cross-reference was fabricated — the reviewer reconstructed
  the pre-edit line from the diff, confirmed it was already one physical line and already checked,
  and showed the identical untouched phrasing at myflow-finish/SKILL.md:88-89 is flagged by a
  heading-rename mutation. The reformat is harmless; the stated rationale was wrong.
  Reviewer also independently reproduced the pathspec behaviour, the two-commit split end to end,
  and the preserve-session-records second-add justification, confirming all three.
Task 5: fix round 1/5 dispatched (resumed the same implementer): narrow the `git add` permission in
  section 4's dispatch clause, and make section 7 clear those three paths from the index before the
  exclusion add rather than assuming nothing upstream staged them. Plus the two report corrections.
Task 5: fix round 1/5 complete (3 addressed, 0 open), scoped re-review dispatched (model: sonnet),
  diff .superpowers/sdd/fix-round-1.diff. Pre-fix tree reconstructed as a02f414 by reading
  task-5.base into a temp index and applying task-5.diff --cached, so the re-review is genuinely
  scoped to the fix rather than re-reading the whole task.
  The implementer's covering check is the strongest in this run so far: it EXTRACTS the staging
  commands from SKILL.md itself and runs them against an index deliberately polluted by a bare
  `git add -A`, so the test exercises the shipped text rather than a copy of it.
  It also refuted the controller's suggested command with evidence: `git restore --staged -- <paths>`
  fails rc 1 and unstages NOTHING when any named path is absent from index and disk — which
  docs/superpowers/ is on every ordinary run — so it would have left openspec/ staged and silently
  reintroduced the defect. `git rm --cached` was rejected too: on a path tracked in HEAD it stages a
  deletion, putting a deletion of the plan in front of the reviewer. `git reset -q -- <paths>` was
  chosen and the reasoning recorded in the skill text so a future editor cannot simplify it back.
Task 5: re-review — all 3 findings ADDRESSED, new breakage: none (model: sonnet). Re-reviewer
  independently reproduced every command-choice claim on git 2.50.1, including that
  `restore --staged` fatals on an unborn HEAD, and ran the §7 sequence verbatim against a polluted
  index to confirm the index ends implementation-only. Confirmed the exclusion pathspec, unfiltered
  review command, two-commit prUrl exception and preserve-session-records ordering all still hold,
  and that myflow-finish/SKILL.md is absent from the fix diff entirely.
Task 5: minor (deferred): SKILL.md:294-295's guardrail says "Always stage with the exclusion
  pathspec above ... never a bare `git add -A`" without naming the new `reset` step, relying on
  "above" to cover both lines of the block. A future "simplification" could drop the reset.
Task 5: complete (uncommitted, review clean, model: opus; reviewer sonnet; 1 fix round).
Task 6: implementer DONE_WITH_CONCERNS (uncommitted, model: opus). Three guards exit 0,
  test-check-unfinished-work.sh 28/28, Step 4's grep returns no match in SKILL.md (severity words
  untouched in the two reviewer-prompt files).
  FOUND A TIME BOMB IN THE BRIEF (controller's defect, not the implementer's): the brief's Step 1
  wording `**Every finding is recorded as a row** in `.superpowers/sdd/final-review-panel.md`:` is
  exactly the cross-reference shape, so check-references.sh reads that target and demands a heading
  inside it. The file does not exist yet — but /myflow-do CREATES it during its own panel run, i.e.
  partway through this very run, at which point the guard would have started failing. Proven by
  planting the artifact: brief wording -> exit 1 naming SKILL.md:156; shipped wording -> exit 0.
  Shipped `**Every finding is recorded as a row.** The panel record is `<path>`:` — a full stop
  inside the bold makes looks_like_section reject it, which is reflow- and gap-independent. No
  suppression marker used; global constraints forbid them.
  Behavioural proof of the column order, as demanded: the table was lifted VERBATIM out of the
  shipped SKILL.md into a fixture worktree and fed to the real guard — fixed -> CLEAR, fourth cell
  open -> "OUTSTANDING: 1 open finding(s)", withdrawn -> CLEAR, and the word "open" in the NOTE
  cell does not fire.
Task 6: reviewer dispatched (model: sonnet), diff .superpowers/sdd/task-6.diff.

Controller check after that finding: scanned Tasks 7-10's remaining plan text for the same shape.
  The only matches are inside fenced example blocks (which the guard skips) or end their bold token
  with a full stop. check-references.sh is currently clean, so no second time bomb is planted.
Task 6: reviewer verdicts — spec ✅ (every requirement and scenario mapped), task quality: 2 Minors,
  no Critical or Important (model: sonnet). Reviewer independently reproduced ALL FOUR table
  behaviours against the real guard using the table lifted verbatim from the shipped file, and
  reproduced both halves of the time bomb using the guard's own CHECK_REFERENCES_ROOT sandbox hook:
  brief wording + planted target -> exit 1; shipped wording + planted target -> exit 0. It also
  traced the stated terminator through to pipeline.md:485-489 to confirm the every-severity bar is
  not an infinite loop.
Task 6: minor (deferred): the findings block sits before `### Optional slot selection`, which is
  still part of the roster-composition narrative; after it (just before `### Panel re-runs`, where
  findings start being processed) would keep that thread whole.
Task 6: minor (deferred): task-6-report.md:65 cites line 156 for the table-header grep match; the
  real match is 158. Every other line number in the report was confirmed correct.
Task 6: complete (uncommitted, review clean, model: opus; reviewer sonnet).
Task 7: implementer DONE (uncommitted, model: opus). 14 harness assertions green (13 RED first),
  three lint guards clean, test-check-unfinished-work.sh 28/28, and the heading is spelled
  identically in skills/myflow-do/SKILL.md:228 and scripts/check-unfinished-work.sh.
  Behavioural proof extracted the heading and sentinel FROM THE SHIPPED PROSE rather than
  hardcoding them, then drove the real guard: None. -> CLEAR, one bullet -> OUTSTANDING, section
  absent -> OUTSTANDING.
  Best methodological note of the run so far: "guard passed" and "guard never looked" are
  indistinguishable, so it ran a NEGATIVE CONTROL — the seven shipped lines plus one planted stale
  reference — and confirmed the guard fired on the planted line only. Worth borrowing for Tasks
  8-10, which all add references.
Task 7: reviewer dispatched (model: sonnet), diff .superpowers/sdd/task-7.diff (7 insertions).
Task 7: reviewer verdicts — spec ✅, task quality Approved, 1 Minor (model: sonnet). Reviewer built
  its own fixture from the shipped prose and drove the real guard three ways, and independently
  replayed the negative control with its own planted stale reference — one failure, on the planted
  line only. Confirmed the producer/consumer contract matches byte for byte.
Task 7: minor (deferred): SKILL.md:222 and :230 both use the bare verb "Refresh" with materially
  different semantics (preserve-ticked-boxes vs re-derive-the-list) and do not cross-reference each
  other. Misapplication fails safe — a lingering stale item yields OUTSTANDING, never a false CLEAR.
Task 7: complete (uncommitted, review clean, model: opus; reviewer sonnet).
Task 8: implementer DONE_WITH_CONCERNS (uncommitted, model: opus). 45/45 per-scenario checks green
  (3/36 before the edits); three lint guards clean; test-check-unfinished-work.sh and
  test-check-finish-preflight.sh both pass; `PR is confirmed open` no longer appears under skills/.
  Negative control fired on the planted line and only there, across all five shipped reference lines.
  THE BRIEF'S `unverified:` COMMIT SEQUENCE WAS WRONG, and the implementer proved it in a sandbox:
  with openspec/ or docs/manual-test/ already staged — this worktree's actual state, and the state
  an operator leaves by running `git add -A` at the human gate — commit 1 SWALLOWED the plan and the
  guide. It added the same `git reset -q -- <paths>` pass Task 5 established, after which commit 1
  carries the implementation alone. The block's second claim (the unconstrained add picks up the
  preserved records) does hold.
  The brief's own line wrapping would have shipped TWO unchecked references — both Steps 1 and 3
  wrap the bold section name onto a different physical line from the path. Proved with a fixture
  where a deliberately stale section name passes clean under the brief's exact wrapping.
  Also corrected two sentences the split falsified, beyond Step 5's two words: 1.2's "land in the
  same commit as the work they describe" and pipeline.md's "one staging pass then covers the
  preserved files". The "do not harmonise the two orderings" instruction is intact.
Task 8: reviewer dispatched (model: sonnet), diff .superpowers/sdd/task-8.diff.

Controller edit between tasks: tasks.md Task 10 Step 1 widened again — commands/myflow-finish.md and
  commands-claude/myflow-finish.md are a FOURTH and FIFTH copy of the command description, both
  still describing run 1 as one commit with no gate. pipeline.md:62-63 requires every command file
  in both trees to agree with the skill it delegates to, because when they disagree whichever the
  agent reads first wins. Verified only these two are stale; the other command files needed no
  change. Files list updated; brief regenerated.
Task 8: reviewer verdicts — spec ✅ (every scenario mapped in both files), task quality Approved
  (model: sonnet). Reviewer reproduced the commit-sequence failure AND fix across three scratch
  repos (pre-staged-and-modified, brand-new-untracked, nothing-pre-staged), confirming the shipped
  five-line sequence always yields commit 1 = implementation only. It replayed the negative control
  by planting a stale token on each of the five shipped reference lines one at a time and confirmed
  the guard fires at exactly those five and nowhere else. It also verified the diff's blob hashes
  match what is staged, so the reviewed diff is exactly the shipped one.
Task 8: minor (deferred): the three course names and the "Stop leaves nothing staged" wording are
  duplicated near-verbatim across SKILL.md and pipeline.md. Consistent today; nothing mechanically
  keeps them so — check-references.sh verifies a section exists, never that duplicated prose agrees.
Task 8: minor (deferred): SKILL.md:118-119's "rather than assuming everything is already staged" is
  carried over from the pre-split text and does not itself motivate the new reset line; that
  rationale arrives three paragraphs later.
Task 8: the reviewer's one Important — stale commands/myflow-finish.md and
  commands-claude/myflow-finish.md — was assigned to Task 10 mid-review and is no longer unowned.
Task 8: complete (uncommitted, review clean, model: opus; reviewer sonnet).
Task 9: implementer DONE_WITH_CONCERNS (uncommitted, model: opus). Per-scenario checks went
  7 passed/41 failed -> 48 passed/0 failed; three guards clean; both new harnesses plus
  test-check-references.sh pass; negative control 7/7, each shipped reference proven to fire on
  BOTH a planted stale token and a renamed heading.
  THE BRIEF'S `unverified:` REMOTE-DELETE BLOCK WAS WRONG, measured against a scratch bare remote:
  deleting an already-absent remote branch exits 1 with "remote ref does not exist", and a REFUSED
  PUSH also exits non-zero — so the brief's `|| true` would have made an expired credential
  indistinguishable from success. The shipped block captures the output and separates the two on
  git's own message.
  Second defect the brief did not anticipate: in the forge-deleted case the stale
  refs/remotes/origin/openspec/<name> SURVIVES locally, and check-cleanup-complete.sh reads that as
  a leftover — so without a prune the new verification step would report a false leftover on exactly
  the case the requirement calls success. The shipped text prunes it on that path.
Task 9: reviewer dispatched (model: sonnet), diff .superpowers/sdd/task-9.diff.

Controller edit between tasks: tasks.md Task 10 Step 1 widened a third time. Task 9's registry
  requirement says every removal rule is stated ONCE, and Task 9 could only enforce that inside the
  two files it owned. Four outer copies still restate cleanup RULES rather than merely naming what
  run 2 does — commands/myflow-finish.md:15, commands-claude/myflow-finish.md:11, AGENTS.md:156 and
  CLAUDE.md:110, the last two carrying the four gating checks, the ignored-file disclosure and the
  preserved-copy condition verbatim. Added a table naming each, and the distinction that matters:
  naming what run 2 does is fine in a reference table; restating a rule is not, and becomes a
  pointer. Brief regenerated.
Task 9: reviewer verdicts — spec ✅ (all ten scenarios), task quality Approved (model: sonnet).
  Reviewer reproduced every load-bearing claim independently: the three remote-delete exit paths
  (0 / 1 "remote ref does not exist" / 128 unreachable) and confirmed the shipped block separates
  them on git's MESSAGE not the exit code; the false-leftover end to end, showing the LEFTOVER
  clause disappears after the prune; and that a SUCCESSFUL push --delete removes the tracking ref by
  itself, so the prune is correctly placed only in the already-gone branch — a precise fix, not a
  defensive one. Path truncation confirmed against a real worktree at a path with a space.
  Guardrail relocation judged CORRECT, with a stronger argument than the implementer's: SKILL.md
  line 12 forces a pipeline.md load on every invocation including the standalone run-2 session, and
  the relocated statements now sit immediately adjacent to the bash they gate — arguably a better
  "moment of destruction" placement than a bullet in a tail-end guardrail list.
  Registry deviation upheld on stronger grounds too: "without a further prompt" is lifted from
  spec.md's own SHALL text, which outranks a design-time annotation that goes stale on archive.
Task 9: minor (deferred): SKILL.md:271 "step 2.1 already proved it merged" is a PRE-EXISTING
  dangling reference (section 2.1 no longer exists after Task 8's renumbering). Task 10 residue.
Task 9: complete (uncommitted, review clean, model: opus; reviewer sonnet).
Task 10: implementer DONE_WITH_CONCERNS (uncommitted, model: opus). All six steps' before/after
  checks flipped FAIL->PASS; three lint guards and seven harnesses exit 0; setup.sh global against a
  sandboxed HOME exits 0 and installs nothing from scripts/, as predicted; five negative controls
  each exited 1 naming the expected line and restored byte-for-byte with sha256 verified.
  THE BRIEF'S STEP 4 GREP UNDER-REPORTED BY CONSTRUCTION (controller's defect): wrong scan set
  (omitted AGENTS.md and CLAUDE.md, which the guard does scan), blind to references split across
  two physical lines, and blind to the `under` connector. Its count of 2 was coincidentally right
  but its MEMBERSHIP was wrong — it returned myflow-status:94 and myflow-do:71, not
  myflow-start:37, which carries the same defect split across two lines. The implementer replaced
  it with an audit reusing check-references.sh's OWN is_associated/looks_like_section predicates:
  4 in the guard's scan set plus 3 line-split instances. Seven references swept, not two.
Task 10: reviewer dispatched (model: sonnet), diff .superpowers/sdd/task-10.diff (11 files).

Task 10 concern, DELIBERATELY LEFT, needs a decision at handoff: the same defect class survives in
  other shapes — `**Finish contract**` shorthand at five sites in myflow-finish/SKILL.md, intra-file
  `see **X** above/below` in pipeline.md, and myflow-do/SKILL.md:145-146 where **Project
  configuration** is split from its path and goes unchecked. Out of this change's spec scope (which
  required sweeping the `**A** in **B** (path)` shape, now done). Candidate for a follow-up issue,
  and for this change's own `## Known incomplete` section — which is exactly the mechanism this
  change builds.
Task 10: reviewer verdicts — Steps 1, 2, 3, 5, 6 ✅; STEP 4 ❌ Critical (model: sonnet).
  The reviewer reimplemented check-references.sh's predicates itself rather than reusing the
  implementer's audit script, and mutation-proved FOUR more genuinely-unchecked references that
  survive the sweep, none of them in the implementer's disclosed Concern 2 list:
    myflow-status/SKILL.md:16      **States** SHADOWED on its own line by **State file**
    state-file.md:119-120          **State self-heal** split across two physical lines
    state-self-heal.md:27-28       **State file** split across two physical lines
    myflow-status/SKILL.md:132-133 **State self-heal** split across two physical lines
  Each proven the same way: rename the target, observe the guard name other lines but never this
  one, restore, sha256 verified. Two are in a file the task was already editing.
  The reviewer also replayed FOUR negative controls on the rewrites that WERE made and confirmed
  each fires exactly as reported — so the seven rewrites are sound; the population was undercounted.
  It upheld the verbose full-heading quote at myflow-start:37 (the guard does grep -qxF against the
  whole normalised heading, so a short token would silently stop matching) and rated the untouched
  myflow-do command files a Minor, noting pipeline.md's "must agree" clause is scoped to the
  state-transition table rather than arbitrary procedural detail.
Task 10: fix round 1/5 dispatched (resumed the same implementer): fix all four, re-derive the
  population from the guard's own predicates, and VALIDATE THE AUDIT ITSELF BY MUTATION — for every
  reference it calls checked, renaming must fail at that line; for every one it calls unchecked,
  renaming must not. An audit nobody mutation-tested is what produced this finding. Report the true
  population and stop rather than sweeping indefinitely if it is much larger than the eleven known.
Task 10: fix round 1/5 — finding ADDRESSED, new breakage none (re-reviewer, model: sonnet).
  Re-reviewer reproduced all five mutation counts EXACTLY (3/12/9/5/2) against an isolated fixture
  via CHECK_REFERENCES_ROOT, confirmed every one of the seven rewritten sites is named by the guard
  once its target is renamed, and checked the NEGATIVE direction too: reconstructed the pre-fix text
  by reverse-applying the diff and confirmed the old myflow-status:16 line was a genuine orphan
  rather than a latent pass, then confirmed two still-unchecked references are NOT named when their
  targets are renamed. Six of seven hunks are pure rewraps with zero wording change.
  Implementer's own retraction is worth recording: audit-refs.sh HAD printed three of the four
  flagged lines as ORPHANs and it triaged them away as "different shape" without checking what they
  pointed at; it also had no concept of shadowing. The rebuilt audit is mutation-validated in both
  directions (30 heading mutations, 61/61 CHECKED and 31/31 UNCHECKED).
  myflow-status:16 needed more than a reflow: the path was BOLD, not backticked, so it was never a
  path candidate and **States** had nothing to associate with. A one-line variant was rejected for
  resolving only because a closing parenthesis made one gap 3 chars instead of 2 — winning by a
  character is not a shape worth landing.
Task 10: re-reviewer nuance the report got wrong: state-file.md:47 and project-configuration.md:28
  DO carry a path-shaped string, merely unbackticked — the same defect as pre-fix myflow-status:16,
  so a one-line mechanical fix rather than the "no path on the line" policy call the report claims.

OPERATOR DECISION (asked and answered mid-run): fix ALL 25 remaining, both families. Policy calls
  decided — a file MAY cite itself by path (Family B gets its own backticked path on the line), and
  Family C repeats the path, verbosity accepted. The six emphasis-coinciding-with-a-heading cases
  are not references and stay untouched. Expected landing point: 86 checked / 6 unchecked.
Task 10: fix round 2/5 dispatched (same implementer), with the both-directions mutation bar and an
  instruction to report any discrepancy rather than adjust the classification to fit.
Task 10: fix round 2/5 complete — all 25 swept, both families. Population landed EXACTLY on the
  stated target, 86 checked / 6 unchecked, with the 6 being precisely the emphasis cases
  (Rules/Skills in AGENTS.md and README.md, Claude Code/Cursor in pipeline.md). No discrepancy.
  Mutation-validated in both directions across all 30 heading texts: 86/86 and 6/6, every heading
  restored byte-for-byte by sha256. Three guards and all seven harnesses green; setup.sh global
  against a fresh sandbox exits 0.
  One Family B item needed a SHAPE change rather than a path: engineering-principles.md:15-16 had
  three section names on one physical line, so adding one path would have associated all three with
  it and left two SHADOWED — the exact round-1 condition. It became a three-item list.
  Family C description corrected as instructed: "no RESOLVABLE path on the line — either none
  written (8 of 10) or written without backticks (2 of 10)".
Task 10: THIRD FAMILY FOUND, deliberately not fixed, and the reasoning is the interesting part:
  state-file.md:45 says `See **Multi-repo shape** below`, but no such HEADING exists — the target is
  a bold paragraph lead-in at :76. It never entered the audit, which only counts tokens naming a
  real heading. Applying Family B's rule mechanically there would have made the guard FAIL, i.e.
  broken the build. Bold lead-ins referenced as sections are a third, unmeasured family, invisible
  to both the guard and the audit. Candidate for the follow-up and for `## Known incomplete`.
Task 10: fix round 2 re-review dispatched (model: sonnet), diff .superpowers/sdd/fix-round-3.diff
  (9 files) — weighted toward new-breakage in canonical contract prose rather than toward the counts.
Task 10: fix round 2 re-reviewer terminated early on a network error (ENOTFOUND) after completing
  its verification and confirming the worktree byte-for-byte unchanged. Resumed from transcript to
  collect the verdict rather than re-dispatching a cold reviewer and paying for the whole pass twice.
Controller verification while waiting: tree intact (31 staged entries, zero unstaged), all three
  guards clean. THIRD-FAMILY CLAIM CONFIRMED INDEPENDENTLY — state-file.md:45 reads
  "See **Multi-repo shape** below." and :76 is `**Multi-repo shape.**`, a bold paragraph lead-in;
  `grep -rn '^#{1,4} *Multi-repo shape' skills/` finds NO heading by that name. Adding a path to :45
  would make the guard demand a heading that does not exist and fail. Leaving it was correct, and
  the implementer's reasoning holds.
PANEL pass 1, full roster of 7, all on Sonnet. Slots 1 and 3 run as general-purpose agents driving
  the repo's own bug-hunter and security reviewer prompts, because subagent_type bugbot and
  security-review do not exist in this harness — substitution recorded, slots not skipped.
Panel slot 5B (Lens B): 2 Important, 2 Minor. Demonstrated a silent false-clear by reordering the
  panel table's columns, which the spec's wording permits.
Panel slot 3 (Security): 1 CRITICAL, 2 Important, 3 Minor/notes. The critical is the headline
  mechanism failing open on format drift: the status cell is matched as literal lowercase `open`,
  so `Open` or `open — disputed` counts as closed and a live Critical finding passes the gate.
  Also demonstrated path traversal via the PR-controlled change-name in BOTH guards.
  Two items reviewed and explicitly cleared rather than left open: --end-of-options is genuinely
  not needed (every ref built from $NAME carries a fixed literal prefix), and the destructive
  removal sequence is unchanged and still gated by its own four checks before the verifier runs.
Panel slot 1 (Bug hunt): 1 Critical, 2 Important, 2 Minor. Its Critical (missing tasks.md reports
  CLEAR) duplicates Security's third finding at a higher severity; its containment finding
  duplicates Security's second but cites a STRONGER precedent — preserve-session-records.sh:78-83
  already allowlists the name, and its comment records that a glob metacharacter once matched and
  overwrote a DIFFERENT change's preserved record. NEW and significant: the two-commit sequence
  fails with "no changes added to commit" when the operator takes the new gate's own Stop path and
  fixes only the guide, because everything changed lives in the three excluded paths.
Panel slot 2 (Principles/Merged): 1 Important, 2 Minor, no Critical. INDEPENDENTLY REPRODUCED the
  empty-implementation-commit failure and supplied the fix shape. Also examined the guardrail
  relocation and judged it correct — the third independent agent to reach that conclusion — while
  recording the residual cost. Verified the global constraints directly: three guards exit 0, both
  harnesses pass, both scripts bash-3.2-safe, no verdict line on the unreadable path, zero commits.

=== REVIEW PANEL ===
Pass 1: full roster of 7 (primary, bug hunt, principles/merged, security, adversarial, lens B,
  lens C), all on Sonnet, dispatched separately. 22 deduplicated findings: 2 Critical, 9 Important,
  8 Minor (+3 notes). Five defects were raised independently by more than one slot. Three slots
  independently judged the guardrail relocation correct.
Fix wave: ONE subagent (model: opus) with the whole list, per contract. All 22 fixed, none
  withdrawn. It corrected the controller's proposed fix — `set -e` is not a reliable stopper on
  bash 3.2 inside a subshell whose parent has errexit off — and used an && chain instead. The new
  parser immediately caught four rows of the panel record ITSELF carrying unescaped pipes.
Scoped re-review: all 22 ADDRESSED, and reproduced the set -e claim independently. FOUND A NEW
  CRITICAL IN THE FIX: a findings row missing only its LEADING pipe is valid GFM, but the awk
  treated it as end-of-table and never resumed — hiding that row and every row below it, no
  unparseable signal. Worse than the four it replaced: one missing character hides unboundedly many.
Fix pass 2: fixed with the right general rule — what ends the table is now a closed stated list
  (blank line, ATX heading, EOF), never "a line the parser did not recognise", which WAS the defect.
  Boundary pipes optional per GFM; unparseable rows counted and the rows after them still read; a
  status-shaped five-cell line outside a table reported. Cases 4j-4p added, including a COUNT
  assertion so a guard noticing only the first row cannot pass.
  It also re-checked the other two signals' parsers for the same class of bug as instructed, and
  verified signal four across ten shapes.
Residual, reported and deliberately NOT fixed, referred to the panel: the `## Known incomplete`
  scraper's fence toggle has no nesting or indentation rule; a guide with no real section but a
  fenced block containing an indented ``` and an unclosed fenced example reports CLEAR. The fixer's
  argument for leaving it — a partial fix kills that reproduction but leaves the marker-length case
  open, so it would look complete while the class stayed open — is sound, and adjudicating it in
  front of the panel is the right call.
Controller: superseded four design.md decisions in place (never rewritten) with fresh IDs and the
  reason each moved, per the Decisions contract.
OPERATOR DECISION: targeted re-run rather than the contract's automatic full escalation — Security,
  Adversarial and Lens B (the slots that found parser evasions) plus Primary as integration check.
  Asked because a 7-slot re-read of a 3300-line diff to validate a ~20-line awk fix is poor value;
  the contract's "do not ask" is overridden by the operator's standing instruction to ask.
Panel pass 2 (targeted: primary, security, adversarial, lens B) — the parser failed open a THIRD
  time, three more ways, found independently:
  - detached row + malformed cell count = invisible (primary, security, adversarial — three slots).
    Caught in-table as unparseable, silently dropped out-of-table. The two defences do not compose.
  - a findings table written inside a GFM blockquote is invisible once a legitimate table exists
    elsewhere; `> ` merges into the first cell so the header never matches and rows split to 6.
  - is_status()'s `==` is locale-collation-dependent: a zero-width char in the status cell reports
    OUTSTANDING under UTF-8 and vanishes under LC_ALL=C, which is the default in many CI images.
  Plus an Important false positive: the orphan rule flags any five-column table with a state-like
  fourth cell, and FENCING the example does not suppress it — the findings scan has no fence
  awareness. And harness cases 4m-4p assert presence, not counts, so a regression that stopped
  after the first hit would pass green.
  Lens B's verdict was the decisive one and it argued from this repo's own history: the guard
  hand-rolls a GFM table parser in awk to recover one fact from a grammar defined in another file's
  prose; .myflow/project.md records check-plan-provenance taking FIVE panel passes and SEVEN fix
  waves on exactly this pattern before being replaced with a real parser; and
  check-cleanup-complete.sh IN THIS SAME CHANGE already solved the analogous problem with
  machine-readable markers instead of parsing.
OPERATOR DECISION: redesign. The emitting side writes one machine-readable marker line per finding;
  the guard becomes an anchored count; the table parser is deleted outright. Dispatched (opus) with
  the fail-safe cases enumerated, the normative-spec requirement, the real fence model for the
  Known-incomplete scraper, count-based harness assertions, and an instruction to attack its own
  design and report what it tried.
REDESIGN complete (model: opus; resumed once after a connection error, from where it dropped rather
  than restarting — the guard rewrite and fence model were already green).
  Signal three is now an anchored marker count plus a `findings-total:` checksum. The table parser
  is DELETED: cells_of, is_delimiter, is_header, ends_table, is_status, the orphan rule, the
  unparseable accounting and the escaping rule, along with every harness case that existed only for
  them. Harness now 128 assertions. Live record migrated: 40 rows / 40 markers / findings-total: 40,
  parses CLEAR.
  The fence scraper was rebuilt to CommonMark's real rule (marker char + length + <=3-space indent,
  unclosed fence runs to EOF); both previously-reported reproductions flip from CLEAR to OUTSTANDING.
  The header no longer overclaims and records what it deliberately does not model.
  SELF-ATTACK, 28 shapes + 13 mutations, THREE ATTACKS SUCCEEDED and were closed:
    - a FENCED marker stood in for a missing one; identifiers and checksum both matched while a
      finding was open -> CLEAR. Closed by requiring the markers to form one unbroken block.
    - CRLF silently disabled the checksum: `read -r` does not split \r and `[ -ne ]` exits 2 inside
      an `if`, which errexit never sees. Closed with parameter-expansion extraction.
    - one NUL byte voided the identifier check: without -a both id lists came back empty and
      compared equal. Closed with -a. Found by mutation testing, not by reading.
  Mutation testing also corrected two of its own overclaiming comments, and measured that LC_ALL=C
  IS load-bearing for the change-name allowlist: `démo` is rejected under C and ACCEPTED under
  en_US.UTF-8.
  Concerns recorded: a marker that simply lies is irreducible (no document format detects a finding
  never written down); the table Status cell can diverge from its marker (marker governs, dangerous
  direction caught, but table-open/marker-fixed clears and misleads a human reader) — flagged for an
  operator decision; the panel record path is not symlink-checked (pre-existing, needs worktree
  write access); `## Known incomplete work` prefix match (pre-existing).
Verification reviewer dispatched (model: sonnet) — reproduce the three successful attacks,
  independently attack the new design, confirm the parser is genuinely gone and nothing regressed,
  and recommend on the Status column.
FINAL ROUND: duplicate-identifier under-count closed (uniqueness now asserted on EACH side
  independently, naming the reused id; three cases, both branches mutation-tested), and the Status
  column dropped per operator decision — table is now ID | Slot | Severity | Location | Note, with
  the marker as the only statement of state. Propagated to all three definitions: spec requirement
  plus two scenarios, the SKILL.md mirror, and the live record, whose 42 rows were rewritten with
  every non-status cell verified byte for byte against the pre-migration copy (zero mismatches).
  The fixer also caught a latent bug it had introduced itself: BSD `seq 1 0` prints "1 0", so a
  zero-finding panel would have been written with two rows.
FINAL STATE: 42 findings recorded, 40 fixed / 2 withdrawn / 0 open. Three lint guards clean, all
  seven harnesses pass, openspec validates strict, setup.sh sandboxed run exits 0, zero commits.
  Panel totals across the run: 7 slots pass 1, 4 slots pass 2 (targeted, operator-scoped), plus a
  verification pass on the redesign — 12 reviewer runs, 2 fix waves, 1 redesign.
  Staged per the contract in force for THIS run (git add -A). The staging split this change
  introduces governs the NEXT run; dogfooding an unmerged rule would desync the human gate from the
  skill actually installed.
