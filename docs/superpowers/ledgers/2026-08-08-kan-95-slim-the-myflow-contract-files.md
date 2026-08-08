# SDD ledger — plan: /Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files/openspec/changes/kan-95-slim-the-myflow-contract-files/tasks.md

Task 1: complete (uncommitted, review clean, model: sonnet; reviewer model: sonnet)
  - controller fix: added MODIFIED `Every pipeline command prints the tab commands at the start of its run`
    to specs/myflow-handoff-output/spec.md (planning gap found by the implementer), and corrected the
    plan's Task 1 Step 5 expected-match list. Synced to both worktree and main checkout.
Task 2: complete (uncommitted, review clean, model: sonnet; reviewer model: sonnet)
  - move fidelity verified independently: 201-line spans, 1 differing line (deleted `below`), permitted edit 2
  - controller revert: implementer hand-edited live openspec/specs/myflow-handoff-output/spec.md;
    reverted, since the change's delta already carries that repoint and the delta is the archive mechanism
Task 3: complete (uncommitted, review clean, model: sonnet; reviewer model: sonnet)
  - value sets rebuilt independently by the reviewer for both folded blocks: identical, no field lost
  - deviation adjudicated CORRECT: myflow-finish/SKILL.md left unfolded — run-1 block alternates
    on-disk/run-only field-by-field, so no adjacent pair is foldable. Plan Step 4 corrected to say so.
Task 4: fix round 1/5 (2 addressed, 0 open — frontmatter description named a dropped column;
  Guardrails permitted-correction bullet lacked the dead-precondition caveat)
Task 4: complete (uncommitted, review clean, model: sonnet; reviewer + re-reviewer model: sonnet)
  - 3 out-of-brief edits adjudicated correct: Bash(gh:*) grant dropped, detail-view PR bullet,
    Guardrails gh bullet — all stale promises caused by removing the probe, not scope creep
Task 4: minor (deferred): skills/myflow-status/SKILL.md:3 frontmatter description does not name the
  Jira column. Pre-dates this change; untouched by it. For the final panel to triage.
Task 5: fix round 1/5 (2 addressed, 0 open)
  - Critical was the CONTROLLER's planning gap: openspec/specs/myflow-state-machine/spec.md carried three
    live SHALLs about self-heal that neither given authority covered. Controller wrote the delta.
    Proposal corrected to record the error rather than hide it.
  - Important: plan's rewritten placement reason for `Resolving a change's worktrees` was FALSE
    (/myflow-do never cites the scan). Correct fix was a verbatim move into finish-contract.md, its
    only consumer — this change's own "reachable from one command" rule, applied to a section the
    change itself stranded. No stub left.
Task 5: complete (uncommitted, review clean, model: sonnet; reviewer + re-reviewer model: sonnet)
Task 5: minor (deferred): the MODIFIED `The state file carries only fields with a live consumer` also
  reworded the exception-extension paragraph, which named self-heal — necessary, meaning preserved,
  slightly wider than the fix instruction literally named. For the final panel to note.
Task 6: complete (uncommitted, review clean after controller adjudication, model: sonnet; reviewer: sonnet)
  - pipeline.md 36,395 -> 31,151 B. Landed above the ~23,000 projection because nine sections came back
    wholly core. Reviewer confirmed no hollowed section, core stands alone, 6 extractions byte-checked.
  - Important finding ADJUDICATED AGAINST THE SPEC, NOT THE CODE: the delta said extraction leaves
    "a new sentence ... and nothing else". The extractions leave rule + operative mechanism + citation.
    The mechanism is CORE by the behavioural test (an agent cannot follow "both commits are guarded"
    without "staged-changes test, one && chain"), so the spec wording would have forced operative
    content into a file no run loads. Spec amended to bound the KIND of content, not a sentence count;
    scenario and plan Step 3 aligned so Tasks 7-9 do not inherit the wrong instruction.
Task 6: minor (deferred): .serena/ is untracked and not gitignored in either checkout, so every
  `git add -A` re-stages it. Controller unstages per task. /myflow-finish run 1's second commit is an
  unconstrained `git add -A` and WOULD commit it. Needs an operator decision at the gate.
Task 7: fix round 1/5 (4 addressed, 0 open — 1 Critical unauthorized rewrap + 3 argument passages
  left in core, plus a 4th eviction the reviewer named and the controller pressed)
Task 7: complete (uncommitted, review clean, model: sonnet; reviewer + re-reviewer: sonnet)
  - SKILL.md 37,900 -> 35,825 B (5.5%). Rule-dense file; honest yield.
  - ROOT CAUSE worth carrying to tasks 8-9: the one passage that drifted was TYPED FROM MEMORY.
    All 13 others were exact-match edits against already-read text and came through byte-clean.
  - Two disclosed judgment calls adjudicated CORRECT: blockquote slice whose first line carries no
    `> ` marker (the marker belonged to the retained lead sentence), and a preserved 2-space
    list-continuation indent.
Task 8: complete (uncommitted, review clean, model: sonnet; reviewer: sonnet)
  - project-configuration.md 45,284 -> 38,275 B; new appendix 13,225 B. All 18 rows byte-verified by
    the reviewer incl. newline positions; only the two permitted `below` deletions differ.
  - Both protected zones intact: `## standards` containment/entry-form rules, and the workspace
    isolation declaration schema.
  - CONTROLLER MEASUREMENT ERROR corrected: my section tallies used a fence-unaware awk, so the
    "12,799 B ## workspace isolation section" in the brief was never a section — it is a worked
    example inside a fence at line 337. Implementer refused to invent a phantom appendix heading and
    flagged it. Task 9's target re-checked fence-aware: 5 genuine ## headings, none fenced.
Task 8: minor (deferred): .superpowers/sdd/task-8-report.md says the core shrank by 6,999 B; actual
  is 7,009. Report file only, gitignored, not in the diff.
Task 9: complete (uncommitted, review clean, model: sonnet; reviewer: sonnet)
  - workspace-isolation.md 31,079 -> 25,317 B (18.5%); new appendix 11,037 B.
  - Strongest fidelity check of the run: reversing all 32 core edits reproduces the pre-task image
    byte-for-byte. Reviewer added the check that misses — appendix-side byte verification, 39 blocks.
  - Disclosed "deviation" adjudicated as CORRECT application of the positional-reference clause:
    a passage whose neighbour refers to it by position stays with that neighbour.
Task 9: minor (deferred): workspace-isolation-rationale.md carries a tool-equivalence permission
  (sha256sum / openssl dgst -sha256 as alternatives to shasum). Appendix-side, not load-bearing since
  the core's own script suffices. Flagged for awareness only.
Task 10: complete (uncommitted, review clean, model: sonnet; reviewer: sonnet)
  - all 23 budget rows re-anchored to floor(actual x 1.25); reviewer recomputed all 23 independently.
  - fixture repair PROVEN legitimate: reviewer copied guard+harness to scratch, neutered the
    over-budget check, confirmed the harness still fails. Not a weakened test.
  - MEASURED per-run /myflow-do load: 210,481 -> 161,521 B (-48,960, -23.26%).
    Proposal projected ~128,500; missed by +33,021 (+25.7% over projection). Reported unspun.
Task 10: minor (deferred): task-10-report.md says 23.27%; correct value 23.26%. Report file only.
