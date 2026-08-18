# SDD ledger — kan-197-require-mutation-test-for-every-guard

Worktree: /Users/tweety53/Projects/agents-worktrees/openspec-kan-197-require-mutation-test-for-every-guard
Branch: openspec/kan-197-require-mutation-test-for-every-guard
Merge base: a805d29465103b0fa8ce29b8933f31e940aa47c6
Recorded models: implementation=sonnet, reviewPanel=sonnet, panelFix=sonnet
Review panel roster: light
Dispatch bundles: [1] [2] [3] [4] [5]

Ledger deliberately written to .superpowers/sdd/tasks/progress.md, which is the ONLY path
scripts/preserve-session-records.sh reads. KAN-73 wrote it to ledger.md and the first preservation
run silently skipped it; see KAN-192.

## Dispatches

Task 1: complete (commit 58fdc76, model: sonnet, review: pending)
  - scripts/lib/coverage.sh + scripts/test-lib-coverage.sh (22 assertions); harness added to
    .myflow/project.md's ## test list, deliberately not ## lint (a library, not a guard).
  - Functions: coverage_reset, coverage_record, coverage_declare, coverage_is_declared,
    coverage_declared_reason, coverage_report, coverage_verdict.
  - commit-fields guard: exit 0.
  - Ran on real /bin/bash 3.2.57 as well as bash 5.3.

  PARENT MUTATION TESTING (not taken from the report — this change is about proving guards can fail,
  so its own library was held to that standard):
    - undeclared zero  -> exit 1, "bravo: 0 checked, and not declared expected-zero (coverage)"
    - empty corpus     -> exit 1, "no member was ever recorded — this guard checked nothing"
    - declared zero    -> exit 0, renders "solo 0 (declared: invokes no guard)"
    All three verified directly by sourcing the library, not via a harness that could itself be
    vacuous.

  IMPLEMENTER JUDGMENT CALL, flagged rather than buried: tasks.md step 2 said the library follows
  panel-record.sh's three grep disciplines (-a, the rc>1 split, -- before paths). The implementer
  found grep has no genuine call site here — the library holds in-process state, reads no file and
  takes no path — so using grep only to validate a count would either duplicate
  PANEL_RECORD_TOTAL_DIGITS' regex or couple to that file for one constant. It used a shell `case`
  glob instead and documented why two of the three disciplines have no literal call site, while
  keeping the posture they encode (reject invalid input, never coerce to a reassuring default).
  Parent accepts: the plan's wording was literal where its intent was about posture.

  Also established empirically and worth keeping: on real bash 3.2.57 under `set -u`, expanding an
  empty array's VALUES ("${A[@]}") throws unbound-variable, while indices ("${!A[@]}") and count
  ("${#A[@]}") do not. Both files' headers record this.

Task 2: dispatched (model: sonnet)
Task 2: complete (commit df9d5dd, model: sonnet)
  - Per-skill coverage = size of rule 2's POST-DELEGATION REQUIRED SET, not raw symlink count.
    The implementer established this and corrected design.md's guessed numbers (14/17/7 -> 13/16/6).
    The unverified: tag on that block worked exactly as intended.
  - Corpus corrections the plan got wrong: openspec-explore IS a member and needed declaring;
    myflow-contracts is NOT a member (filtered out by name), so declaring it would be inert.
  - Declared: myflow-start, myflow-status, openspec-explore -- each with its own reason.
    myflow-fast deliberately NOT declared; its required set resolves to 16 through delegation.
  - PARENT MUTATION TESTING: collapsing myflow-finish's required set to zero -> exit 1,
    "myflow-finish:0: 0 checked, and not declared expected-zero (coverage)". Tree restored clean.
  - PARENT-FOUND LIMITATION, now in design.md: a PARTIAL collapse is silent. Stripping most of
    myflow-do's citations took its required set 13 -> 5 and the guard still exited 0, correctly per
    the rule as specified. Catching that needs the per-member baseline rejected under
    declared-expected-zeros (which would not have caught KAN-73 at all). Recorded rather than
    scope-crept; follow-up candidate at finish.

Tasks 3, 4, 5: complete (commits b099f53, 7ab98e5, 45e65d7, model: sonnet)
  - commit-fields guard: exit 0 on all three. All four guards exit 0 on the real tree.
  - Task 4 corpus is per-TARGET (8), not per-file (160): collect_hits batches per target with no
    per-file skip logic, so a per-file count would be non-zero by construction -- ceremony, which
    the design explicitly warned against. Good architectural judgment by the implementer.
  - Task 4 has ZERO declared members (EXPECTED_ZERO_TARGETS=()) -- all 8 targets non-empty, recorded
    as a measured finding rather than an omission.
  - Task 5 flipped existing cases 10 and 14, which tested "no marks at all" -- a scenario the new
    rule now correctly flags. Not a weakening; the old expectation encoded the defect.
  - check-stage-mark-calls.sh has no CHECK_*_ROOT sandbox override (unlike its siblings), so its
    declared-zero shape is proven against the real repo rather than a fixture. Weaker isolation than
    the other three; worth a panel opinion.
  - No synthetic declarations were invented anywhere to silence a real finding -- the implementer
    renamed fixtures to legitimately-declared paths instead. Verified.

PARENT CONCERN FOR THE PANEL -- the mechanism's signal strength varies sharply by guard, and the
declaration volume is the tell:
    check-vocabulary        0 of 8 declared   -> every member checks something; strong signal
    check-guard-symlinks    3 of 6 declared   -> a zero is anomalous; strong signal
    check-references       29 of 59 declared  -> a zero is ORDINARY for a Markdown file with no
                                                 cross-reference; weak signal, long list
    check-stage-mark-calls 27 of 33 declared  -> a zero is ORDINARY for a skill file with no marks;
                                                 weak signal, long list
  Verified the lists ARE written path entries, not runtime-inferred, so the spec's "never inferred
  from the tree" holds. The open question is whether a 29-entry declaration list is reviewed or
  merely skimmed -- and whether "this file has no references" is the same KIND of fact as
  "this skill requires no guards", which was anomalous. Put to the panel deliberately rather than
  settled by the parent.
