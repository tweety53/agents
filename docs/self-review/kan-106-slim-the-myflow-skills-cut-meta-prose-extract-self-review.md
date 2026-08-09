# Self-review — kan-106-slim-the-myflow-skills-cut-meta-prose-extract

**Rating:** 3/5 (operator)

## Problems and fixes

- **`tasks.md` authoring bug: duplicate `**Files:**` fields.** Two tasks (8 and 10) had a second,
  single-line `**Files:**` field right before their `Tests:`/`Regression:`/`Baseline:`/`Commit:`
  block, duplicating the metadata block's own `**Files:**` bullet list earlier in the same task.
  `check-task-commit-fields.py`'s field parser takes the *last* `**Files:**` occurrence, so the real
  file list was silently discarded and replaced by the short summary line — the guard then failed
  both tasks with "file not declared," even though the implementer had touched exactly the right
  files. Root cause: I template-drafted the field family (`Files/Tests/Regression/Baseline/Commit`)
  by pattern-matching a nearby example that had this shape, without noticing the earlier bullet-list
  `**Files:**` in the same task already existed. Fixed by removing the duplicate line in both tasks.
  **Fix for next time:** grep every task body for `^\*\*Files:\*\*` before dispatching an
  implementer — a count greater than one per task is the defect, and it's a five-second check that
  would have caught both instances before they ever reached the guard.
- **Same root cause, milder form, three more times.** Tasks 2, 4, 8 (again) also hit the guard's
  `Files:` check on first pass — not from a duplicate field, but because the field's prose
  ("and every other `skills/myflow-contracts/*.md` core found to repeat the boilerplate... located
  by Step 1's grep, not pre-enumerated") described a dynamically-discovered set rather than naming
  it. The parser only picks up backtick-quoted tokens; prose describing "everything Step 1 finds"
  contributes nothing. I fixed each by updating the `Files:` field with the real, concrete list once
  the implementer reported what it actually touched — correct, but reactive. **Fix for next time:**
  for a task whose file set is genuinely discovered at runtime, write the `Files:` field as "TBD,
  update from the implementer's report before the commit-fields guard runs" rather than a
  half-descriptive sentence that reads like a real declaration but isn't backtick-parseable.
- **A backtick around a script name in a `Tests: none` field was read as a declared test.** Task 1's
  `Tests:` field read "...resolving via `check-references.sh`." — the backtick-quoted script name
  triggered the parser's fallback-to-backtick-tokens path (used when no `Case <N>` label is present),
  so the guard expected `check-references.sh` to appear in the commit's diff as if it were a test
  name. Fixed by removing the backtick quoting from that one mention. **Fix for next time:** never
  backtick-quote a filename inside a `Tests: none` field's explanatory prose — say "the
  check-references guard" instead, or keep the field to the word "none" and put any pointer in the
  step's own Verify block instead.
- **The final review panel found 6 real defects, 5 of them in code I never independently verified
  by hand** — the awk-parser duplication (F1), a genuine octal-parsing bug in port arithmetic (F3), a
  missing `-x` check (F4), and a silent-empty-string bug in `resolve_value_refs` (F5) were all in
  `scripts/prepare-workspace.sh`, and every one of them survived that task's own per-task review
  clean. The per-task reviewer read the script for structure and cross-checked it against the worked
  example in `workspace-isolation.md`, but didn't independently exercise edge cases like a
  leading-zero default or a non-executable guard file. **Fix for next time:** for any new script
  handling untrusted/declared configuration values (table cells from a project file, in this case),
  explicitly instruct the per-task reviewer to enumerate edge cases in the input grammar (leading
  zeros, missing permissions, forward references) rather than trusting a structural read plus the
  happy-path test cases the implementer itself wrote — the implementer's own tests will only ever
  cover what the implementer thought to worry about.
- **Task 11 (compress the two reviewer prompts) was flat wrong on its own premise**, and it took a
  whole-branch panel re-check to catch it — the per-task review for Task 11 itself found nothing,
  because it checked "was the negation/exclusion content preserved" (yes) rather than "should this
  file have been compressed at all" (no). The ticket's own justification — "a reviewer's output is
  chat, not an artifact, so the caveman Boundaries rule already permits this" — conflates the
  *subagent's output* (genuinely chat) with the *prompt template file* (genuinely persisted,
  repo-committed, read by every future run). I wrote the design doc's Item H section myself and
  reproduced that same conflation into the plan without catching it, and the per-task reviewer I
  dispatched inherited the same framing and checked only what the framing asked it to check.
  **Fix for next time:** for any task whose entire premise is "X is fine to do because of rule Y,"
  explicitly ask the reviewer to evaluate whether rule Y actually applies to X — not just whether
  the task executed rule Y correctly once granted the premise.
- **Two agents ran `git rebase --autosquash` against the same worktree in parallel** (fixing F2 and
  F3-F5 at the same time) — a real risk of corrupting the branch, since git rebase operations aren't
  safe to interleave in one working tree. It happened to serialize cleanly this time (verified by
  reading the resulting history afterward), but that was luck, not design. **Fix for next time:**
  never dispatch two fix-round agents against the same worktree in parallel when either might run
  `git rebase`, `git commit --amend`, or any other history-rewriting operation — serialize those,
  even when the fixes touch disjoint files, because the git operation itself (not the file content)
  is what collides.

## Cost

13 implementer dispatches (one per task, all Sonnet), 13 initial per-task reviews, 4 fix rounds with
matching re-checks during per-task review (Task 2: 1, Task 6: 1, Task 10: 1), plus the whole-branch
final panel: 3 slots × 2 full passes (pass 1 + re-check 1) + 2 slots × 1 more pass (re-check 2, after
the Task-11 revert) = 8 more dispatches, plus 2 fix-round dispatches for the panel's own findings
(F1/F3-F5 combined, F2, F6) = 3 more. Total: roughly 13 + 13 + 8 (per-task fix/re-check pairs) + 8
(panel passes) + 3 (panel fix rounds) + 1 (revert dispatch) ≈ 46 subagent dispatches across the
whole run, all on Sonnet per the recorded `models.implementation`/`models.reviewPanel` choice. This
is a large change (13 planned tasks, 26 files, ~2570 changed lines in the final merged diff) applying
an established mechanism (KAN-95's partition rules) rather than inventing one, which is why the
per-task dispatch count scales roughly 1:1 with the task count rather than needing extra
exploration rounds.

## What went well

- **The `myflow-contract-economy` machinery (rule extraction, per-move ledger, the bounded
  carve-out) worked exactly as designed** across 5 different prose-slimming tasks (1, 2, 4, 5, 9) —
  every implementer correctly distinguished "move verbatim" from "extract a rule," and every
  per-task review caught real byte-fidelity violations when they occurred (Task 2's fix round found
  4 genuine drift issues: a duplicated lead-in, a paraphrased sentence, a doubled sentence, a
  capitalization change) — the mechanism is sound, and reviewers actually used it as a checklist
  rather than skimming.
- **The final whole-branch panel earned its keep.** Every one of its 6 findings was real: not a
  single false positive across two independent reviewers converging on the same duplication issue
  (F1), a genuine content-loss catch (F2) that the per-task review for Task 4 missed, three real
  bugs in a new script (F3-F5) that no per-task review had exercised, and a correctly-identified
  premise error (F6) in an already-"clean" task. This is a strong argument for running the
  whole-branch pass even when every individual task passed its own review — integration-level
  problems are a different failure class from per-task problems, and this run demonstrated why both
  gates exist.
- **Recovering from my own plan-authoring bugs (duplicate `Files:` fields, prose instead of
  backtick lists) was fast and mechanical** once identified — each was a one-line `tasks.md` edit,
  re-run the guard, confirm exit 0, continue. The guard's error messages were specific enough
  (naming the exact undeclared file, or the exact phantom test name) that no investigation was
  needed beyond reading the message.

## Automation candidates

- **A `tasks.md` self-check before dispatching the first implementer** would have caught 4 of the 5
  `Files:`-field defects before they cost a guard failure and a manual fix cycle: grep every task
  body for more than one `^\*\*Files:\*\*` line (the duplicate-field bug), and flag any `Files:`
  field whose value, once backtick-tokens are extracted, is empty or clearly smaller than its prose
  implies (the "described but not declared" bug). This is exactly the kind of mechanical check
  `scripts/check-task-commit-fields.py` already knows how to do against a real commit — running an
  equivalent check against the *plan* before any commit exists would move the failure earlier and
  cheaper.
- **A lint pass for backtick-quoted filenames inside a `Tests: none` field** — since the parser's
  own fallback behavior (treat any backtick token as a declared test when no `Case <N>` label is
  present) is now a known footgun, a one-line grep (`grep -B1 "Tests:.*none" tasks.md | grep '\`'`)
  run once before dispatching implementers would catch this class of authoring mistake for free.
- **The two agents' concurrent `git rebase --autosquash`** worked by luck; a mechanical safeguard —
  a simple lockfile in the worktree that a fix-round dispatch prompt checks for and refuses to
  proceed if held — would convert "serialize by hand, hope nothing races" into an actual guarantee,
  for the specific and recurring case of dispatching more than one fix agent against the same
  worktree in the same turn.
