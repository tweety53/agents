# Self-review — kan-265-be-brief-in-repo-markdown

**Jira:** KAN-265 — Apply the be-brief rule to every Markdown file in the agents repo
**Command:** `/myflow-fast`, one creating run carried through to `FINISHED`
**Rating:** 3 of 5 — acceptable. The verification is the strongest part: a normative-sentence
inventory held byte-identical through ten commits and was re-proved against a moved base. What
holds it to a 3 is that the per-task review was skipped on 8 of 11 tasks, four plan defects reached
dispatch, the corpus was mismeasured three times, and a guard shipped hours earlier broke `## lint`
in every apply worktree.

Context bundle: 4 of 4 sources found.

## Angle 1 — Problems encountered — `myflow-fix`

- **[myflow-fix]** A task can be marked complete with no per-task review, and nothing records or checks it. The review ran for 3 of 11 tasks; `check-task-commit-fields.sh` knows nothing about reviewers and the parent marks the checkboxes, so eleven tasks looked complete. Each of the three reviews that did run found a Major, one of them in an edit the parent had approved. — filed: KAN-281
- **[myflow-fix]** Four plan defects survived writing-plans' self-review and three guards: a task declaring a commit COMMIT-PER-TASK forbids, `Files:` fields written as globs the commit-fields guard silently ignores (verifying nothing for eight tasks), and two tasks omitting files the work necessarily touched. — filed: KAN-282
- **[myflow-fix]** A guard's own harness cannot catch a defect that only appears in a worktree. `check-installed-rules.sh` failed in every apply worktree, and the first fix made eleven of its own cases pass vacuously because the harness runs from a worktree too. — filed: KAN-283
- **[myflow-fix]** The corpus was mismeasured three times, always by sweeping the whole tree instead of the enumerated roots. `check-normative-inventory.sh` printed "92 file(s)" from its first run; the error was reading past it. No pipeline change would fix an operator misreading its own guard. — declined

## Angle 2 — Token and time cost — `myflow-cost`

- **[myflow-cost]** The base moved three merges ahead mid-change and the conflict surfaced only at `gh pr merge`. Because the safety argument is a comparison against a recorded baseline, and the corpus itself had moved, every proof had to be re-derived. Two genuine defects were also only found then. — filed: KAN-285
- **[myflow-cost]** Two tasks yielded almost nothing — 6.1 cut 102 bytes from 462 KB, and `openspec-explore` cut nothing at all. Both were correct outcomes and both cost a full dispatch to establish; "read it and found nothing" is a result worth paying for. — declined

## Angle 3 — What went well — `myflow-improvement`

- **[myflow-improvement]** The inventory guard justified itself by what the panel found inside it: two Majors, both the same class — a corpus that could be silently emptied while the guard reported success. Neither was reachable from any single task's diff. — declined
- **[myflow-improvement]** Every implementer escalated rather than deciding alone — the precedence sentences, the per-citation tails, the count gap, the uncovered `skills/README.md`. Two of those escalations became tasks that would not otherwise exist. — declined
- **[myflow-improvement]** Cuts-only held under pressure. Seven rotted copies were found and every one was cut rather than corrected, because correcting is a reword. That discipline is what kept the inventory byte-identical through ten commits. — declined

## Angle 4 — Automation — `myflow-automation`

- **[myflow-automation]** The two automation candidates worth building are a guard that fails when a task is checked with no recorded review, and either rejecting globs in `Files:` or expanding them there. Both are stated in KAN-281 and KAN-282 rather than filed twice. — declined
- **[myflow-automation]** The empty-comparison trap — an exit-2 run leaves empty stdout, and two empty files compare equal — belongs in the shared dispatch preamble. It was passed into every later prompt by hand after the parent hit it. — declined

## Angle 5 — Go app and persistent storage — `myflow-stats-app`

- **[myflow-stats-app]** The store rejects a same-state write as "moving backwards", so a fix run cannot write state back — which **A fix never moves the state** requires it to do. `prUrl` and a corrected `mergeBase` were both unrecordable until the terminal transition. Filed together with `mergeBase` accepting a path where a sha belongs, since the two compound: the value was wrong and the store would not take a correction. — filed: KAN-284

## Outcome

Five issues filed — KAN-281, KAN-282, KAN-283 (`myflow-fix`), KAN-284 (`myflow-stats-app`),
KAN-285 (`myflow-cost`) — all labelled `myflow`, `AI-generated` plus their angle label.

Seven further findings this change could not fix are recorded in the archived change's
`followups.md`, including two live requirements that contradict each other about when commits
happen, and always-on text still telling every session to review a staged diff that `/myflow-do` no
longer produces.
