# SDD ledger — kan-73-install-guard-scripts-alongside-skills

Worktree: /Users/tweety53/Projects/agents-worktrees/openspec-kan-73-install-guard-scripts-alongside-skills
Branch: openspec/kan-73-install-guard-scripts-alongside-skills
Merge base: 7922242093bc46e57434fe2a18b44634930b292c
Workspace id: kan-73-019c (computed; no workspace created — no task starts the daemon)
Recorded models: implementation=sonnet, reviewPanel=sonnet, panelFix=sonnet
Review panel roster: light
Dispatch bundles: [1] [2] [3 4] [5] [6]

## Dispatches

Task 1: complete (commit 21c45b4, model: sonnet, review: combined spec+quality — dispatched, model: sonnet)
  - commit-fields guard: exit 0. Files/Tests/Commit checked against git and matched.
    Regression and Baseline reported "skipped, not verified" — the project's `## test` command
    cannot target a single named test. Documented guard behaviour, not a failure.
  - assertion counts: test-plan-dispatch-bundles 19 -> 21; test-check-workspace-isolation 146 -> 149.
  - PLAN DISCREPANCY reported by the implementer: tasks.md Task 1 Step 1 claims
    `scripts/test-plan-dispatch-bundles.sh` "already exercises the PLAN_DISPATCH_BUNDLES_ROOT
    override". It does not — zero references to that variable existed in the 236-line file at
    7922242. The claim was wrong when written. It did not affect the implementation, because the
    new case needed the override's *absence*, which was the harness's default state anyway.
    Recorded here rather than silently dropped.

Task 2: dispatched (model: sonnet)
Task 2: complete (commit 478cecd, model: sonnet, review: pending dispatch)
  - 39 symlinks: myflow-do 14, myflow-fast 17, myflow-finish 7, myflow-status 1.
    myflow-start and myflow-contracts correctly have none.
  - Parent verified independently: all 39 are git mode 120000, all resolve, none absolute.
    Sample stored target: ../../../scripts/check-unfinished-work.sh
  - commit-fields guard FAILED TWICE on parent-authored plan defects, not on the commit:
      1. Task 2's `Files:` declared four directories; the guard checks individual paths. Rewritten
         to the 39 real paths.
      2. Task 2's `Tests:` named test files belonging to tasks 5 and 6. Rewritten to declare none.
      3. First rewrite still failed: prose + a blank line after `**Files:**` terminated the field
         before the bullet list. The list must follow `**Files:**` directly.
    Third run: exit 0. No fix-round slot consumed — a guard failure is not a review finding, and
    the defect was in the plan's declaration rather than in the implementation.
  - Implementer flagged a zsh word-splitting gotcha in the plan's Step 1 idiom (unquoted `$VAR`
    does not word-split in zsh); it self-corrected and redid creation under bash. Worth a note for
    anyone reproducing Step 1 by hand.

Parent correction before bundle 3: Task 7 step 2's precomputed budget figure (23910) was stale --
tasks 3 and 4 also grow skills/myflow-fast/SKILL.md. Step 2 now measures the final size instead.

Bundle 3 (tasks 3, 4, 7): dispatched (model: sonnet)

Task 1 review: complete (model: sonnet, review: combined spec+quality)
  Verdict: no Critical, no Major. One Minor, open.
  - MINOR (open, queued): scripts/plan-dispatch-bundles.sh — resolve_file now runs unconditionally
    even when PLAN_DISPATCH_BUNDLES_ROOT is set. Bash's ${VAR:-expr} evaluates expr lazily, so the
    pre-change code never ran the fallback when the override was set; the new code does. Verified by
    the reviewer with `bash -x`. Impact is small but it is an unrequested behaviour change: a caller
    setting the override to sidestep a broken invocation path would now still hit exit 2 from the
    resolver. FIX DEFERRED until bundle 3's implementer is out of the worktree — the fix is a
    --fixup into 21c45b4 plus a --autosquash rebase, which must not run concurrently with a commit.
  Reviewer verified independently and positively:
  - "ported verbatim" claim TRUE, diffed byte-for-byte across all four copies of resolve_file.
  - Both new test cases reproduce the regression against the real pre-fix code at 7922242.
  - Loop correctness: multi-hop chains, relative targets, 40-hop cycle bound, cd failure, spaces in
    paths, bash 3.2 portability — no defects.
  FOLLOW-UPS (not defects in this task; candidates for a follow-up issue at finish run 1):
  - Extract resolve_file into scripts/lib/resolve-file.sh for plan-dispatch-bundles.sh,
    preserve-session-records.sh and gather-self-review-context.sh — 4 copies now exist and
    lib/panel-record.sh's header records drift between duplicated helpers as exactly why it exists.
    Deliberately EXCLUDE check-workspace-isolation.sh: its own comment (and check-cleanup-complete.sh's)
    states it is single-file by design, copied into projects one at a time, so a sourced helper would
    make it present but unrunnable. Reviewer's recommendation, and it is well grounded.
  - Test coverage gap in the ported helper, pre-existing and shared by all copies: no fixture exercises
    a symlink cycle or an absolute symlink target, so deleting the 40-hop bound or breaking the
    absolute-target branch would pass the suite.

Bundle 3 (tasks 3, 4, 7): complete (model: sonnet)
  Original shas 6f06ecd / 4738325 / 8806530, rewritten by the fixup rebase below.
  Task 4 flagged by its own implementer: its commit left skills/myflow-fast/SKILL.md at 18478
  against an 18225 budget, i.e. check-contract-budget.sh RED at that commit, clean only after
  task 7. Build-green contract says `green` means "this task's own verification command can run",
  and task 4's step 4 verification includes check-contract-budget.sh -- so the tag was violated.
  Cause: the PARENT's dispatch briefing said "only task 7 touches that row, using the final size",
  overriding task 4's own step 4 ("raise that one row again"). Parent error, not implementer error.

Fix round 1 (parent-applied, no implementer dispatched -- both fixes were parent-authored):
  - fixup into task 1: make resolve_file lazy again in scripts/plan-dispatch-bundles.sh, closing
    the Minor from task 1's review. Verified with `bash -x`: 0 resolve_file trace lines when
    PLAN_DISPATCH_BUNDLES_ROOT is set. Guard still resolves correctly through the symlink farm.
  - fixup into task 4: raise skills/myflow-fast/SKILL.md's budget row to 23097 (18478 x 1.25) so
    task 4's own commit is green; task 7 then raises it to 24613 (19691 x 1.25). Two real raises,
    both commits green.
  git rebase --autosquash: two expected conflicts on check-contract-budget.sh's single row
  (23097 vs 18225, then 24613 vs 23097), both resolved to the later value. Rebase clean.

  Rewritten shas: task 1 b45cf19, task 2 94c04c2, task 3 51d781c, task 4 ceff369, task 7 ec36304.

  Per-commit budget compliance verified AFTER the rebase -- every commit green:
    b45cf19  fast 17915/18225   pipeline 41853/44574
    94c04c2  fast 17915/18225   pipeline 41853/44574
    51d781c  fast 17915/18225   pipeline 43264/44574
    ceff369  fast 18478/23097   pipeline 44583/55728
    ec36304  fast 19691/24613   pipeline 44583/55728
  All nine in-scope lint guards exit 0 on the final tree.

  commit-fields guard failed again on two more PARENT-authored plan defects, same class as task 2:
    - task 3's and task 4's `Tests:` fields named guard scripts (check-references.sh,
      check-contract-budget.sh, test-check-guard-symlinks.sh) as though they were tests carried by
      those commits. Rewritten to declare none.
    - task 4's `Files:` did not declare scripts/check-contract-budget.sh, which the fixup made it
      touch. Added.
  All five commit-fields guards now exit 0.

Bundle 4 (task 5): dispatched (model: sonnet)
Review of tasks 2, 3, 4, 7: dispatched (model: sonnet)

Task 5: complete (commit 368e3b2, model: sonnet, review: pending)
  - scripts/check-guard-symlinks.sh (578 lines) + scripts/test-check-guard-symlinks.sh (343 lines),
    wired into .myflow/project.md's ## lint and ## test lists.
  - Real tree: GUARD-SYMLINKS-OK -- 53 guards across 6 skills validated, exit 0.
  - 9 fixture trees, 32 assertions. Ran under bash 5.3 and the machine's real /bin/bash 3.2.
  - Rule 3's classifier is SHAPE-BASED, with no whitelist of guard names: a scripts/<name> citation
    counts as an invocation only inside a ```bash/sh/zsh fence on a non-comment line, or when
    immediately preceded by Run/Invoke/Execute outside a fence. It independently reproduces all
    three borderline cases as prose. Fence tracking honours backtick length, so a nested 4-backtick
    example fence does not desync it.
  - Rule 2 derives each guard's sibling dependencies by grepping that guard's own source for
    $SCRIPT_DIR/<name>, never from a copy of design.md's map, so a new sibling is covered
    automatically.
  - Implementer discovered and documented: macOS system awk does not honour `--` as an
    end-of-options marker (unlike grep). Fixed by feeding files via stdin redirection rather than
    relying on a marker awk ignores.
  - Implementer scoped rule 2 to each skill's OWN directory after an early design that followed
    cited contract paths produced false positives on myflow-status and myflow-start. Documented in
    the guard's header. Trade-off recorded: a guard invoked solely inside a shared contract's
    usage-synopsis prose is not detected.

  PARENT VERIFICATION (mutation testing, not taking the report at face value):
    - clean tree -> exit 0
    - replace a symlink with a real file copy -> exit 1, names the path and cites rule 1
    - remove a symlink entirely -> exit 1, names the invoking SKILL.md line and cites rule 2
    Tree restored; git status clean afterwards.
    NOTE: an earlier reading of "exit=0 on a violation" was a measurement artifact of piping the
    guard through `head` and reading head's status. The guard's own exit codes are correct.

INCIDENT: the worktree's UNTRACKED planning artifacts (openspec/changes/kan-73-.../ and the
  docs/superpowers/specs/ design doc) were deleted during task 5, almost certainly by a `git clean`.
  Nothing was lost -- the main checkout is authoritative and they were re-copied -- and
  .superpowers/sdd/ledger.md survived because .superpowers/ is gitignored. Task 6's dispatch now
  carries an explicit prohibition on `git clean` and on deleting untracked files.

Task 6: dispatched (model: sonnet)

Review of tasks 2, 3, 4, 7: complete (model: sonnet, review: combined spec+quality)
  Verdict: no Critical. 1 Major, 4 Minor -- all OPEN, fix round queued behind task 6.

  MAJOR 1 (open): task 4 added a **Guards:** field to five handoff blocks across myflow-do,
    myflow-finish (both blocks) and myflow-fast, but never added it to
    skills/myflow-contracts/handoff-blocks.md, which is CANONICAL for those templates and states at
    line 24 that a field a command emits and the template omits "is drift the moment /myflow-status
    renders the same state". Failure: /myflow-status regenerating either block omits **Guards:**,
    silently disagreeing with what the producing command printed for the same run.
    PARENT VERIFIED INDEPENDENTLY: handoff-blocks.md contains zero occurrences of "Guards"; the
    cited sentence is really at line 24; three skills really emit the field.

  MINOR 2 (open): pipeline.md:417-423's six-name prose exemption reads exhaustive but is not.
    Other non-invoking mentions left repo-relative and uncovered by the list: myflow-do/SKILL.md
    :438,:450,:596; finish-contract.md:244,:333; myflow-finish/SKILL.md:30; pipeline.md:172,:211.
    Nothing executes wrong -- the reviewer checked each in context and task 5's independent
    shape-based classifier agrees they are prose. Documentation precision: the rule undersells its
    own scope and could mislead a future editor trusting the six names as complete.

  MINOR 3 (open): task 3's declared Files: lists myflow-status/SKILL.md, myflow-fast/SKILL.md and
    check-contract-budget.sh, none of which 51d781c touched. The commit-fields guard does not catch
    this (it checks diff-files-subset-of-declared, not the converse). Plan/commit mismatch.

  MINOR 4 (open): task 4's presence check covers only guards a skill's own text names, not the
    siblings a named guard resolves internally (prepare-workspace.sh -> check-workspace-isolation.sh,
    check-task-commit-fields.sh -> .py, plan-dispatch-bundles.sh -> .py). A broken sibling symlink
    would not trip GUARDS MISSING; it would surface as that guard's own runtime failure at the
    moment it is needed -- which the hand-run-fallback narrative does not anticipate. The stated
    guarantee reads stronger than what is delivered.

  MINOR 5 (open): myflow-status/SKILL.md:26 declares a presence check for check-finish-preflight.sh,
    but :114 says its merge-status logic is deliberately NOT re-derived from that script. The skill
    never invokes it, so "guards it can invoke" does not literally apply.

  Reviewer confirmed clean, with evidence: resolution rule stated exactly once (pipeline.md:403-423);
  all 39 symlinks relative and mode 120000 and resolving; every budget raise equals actual_size x 1.25
  floored at the commit that raised it, with no other row touched; task 7's carried prose byte-identical
  to both backup copies and to what landed in ec36304; no gating language in any of the four commands.

  REVIEWER INCIDENT: ran `git checkout 51d781c -- .` in the worktree by mistake, then
  `git reset --hard HEAD`. No work lost; HEAD correct. It left an empty stash entry behind, which
  the parent has now dropped. Second agent in this run to disturb the worktree's state -- both
  recovered, but reviewers should be read-only and this one was not.
