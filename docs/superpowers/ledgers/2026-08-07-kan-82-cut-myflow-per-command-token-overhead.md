# SDD ledger — kan-82-cut-myflow-per-command-token-overhead

Recorded models come from the state file: `models.implementation` = Sonnet,
`models.reviewPanel` = Sonnet, `models.panelFix` = Sonnet. All three are operator choices recorded
at the creating `/myflow-start` run, overriding the pipeline defaults of Opus / Sonnet / Opus.

## Task 1.1 — move `Follow-up issues` into `jira-followups.md`

- Implementer dispatch: general-purpose, model: **sonnet**. TASK_BASE `adedf66`.
- Review dispatch: general-purpose, model: **sonnet**. Diff:
  `.superpowers/sdd/tasks/review-adedf66.diff`.
- Result: complete (uncommitted, review clean after one operator decision, model: sonnet).

### Review findings and their resolution

| ID | Severity | Location | Resolution |
|----|----------|----------|------------|
| T1-1 | Major | `skills/myflow-contracts/jira-integration.md:166`, `:257` | withdrawn by the operator |
| T1-2 | Major | `skills/myflow-contracts/jira-followups.md:8` | not a defect — the plan clause behind it was struck |

**T1-1 — withdrawn by the operator**, with the reason they gave: the two citations read
`**Follow-up issues** (`jira-integration.md`) below`, and a real section by that name **is** below in
that file — the forwarding stub — so nothing false is stated; the cost is one extra hop for the
reader. Repointing them would have made the word "below" false instead, and the operator's round-3
planning decision permits repointing a path but not editing a connecting word.

**T1-2 — the implementer was right and the plan was wrong.** Task 1.1 offered to promote the moved
`### Follow-up issues` heading to `##`; the reviewer read that clause as satisfied, the implementer
declined because promoting breaks the task's own zero-lost-lines check. The delta spec settles it:
citation repointing is the split's only permitted edit, and a heading-level change is not one. The
clause was struck from `tasks.md` and a global constraint added so tasks 2.1 and 3.1 cannot hit it
again.

**Two further plan defects were corrected in the same pass**, both mine and both found through this
review: tasks 1.1, 2.1 and 3.1 each expected "nothing at all as `<`" from the line-multiset check,
which forbade the one edit the split permits — `design.md`'s verification table had the exclusion and
the tasks dropped it. Task 2.1 also now states why a moved line produces no diff output at all, which
is the fact that makes a `<` line evidence of loss rather than of movement.

## Tasks 4.1, 5.1, 6.1 — performed by the parent session

Dispatched to no subagent. Model: **opus** (the parent `/myflow-do` session). Recorded because the
model policy's audit trail is per dispatch, and a step the parent performed itself is still a step
whose model someone may later want to know.

## Review panel

Every slot in both passes: model **sonnet**, from `models.reviewPanel`. Slots 1 and 3 were run as
`general-purpose` subagents reading `bug-hunter-reviewer-prompt.md` and
`security-reviewer-prompt.md`, because the `bugbot` and `security-review` agent types do not exist
in this harness — so their entries read `sonnet (prompt-file substitute)` rather than
`unknown (agent-defined)`, the dispatcher having chosen the model in this run. The full roster, the
selection reasons and the deviation are in `final-review-panel.md`.

The markdown fix round: model **sonnet**, from `models.panelFix`. The shell fix rounds were
performed by the parent session on **opus**.

**This file was written to `.superpowers/sdd/ledger.md` during `/myflow-do` and moved here at finish
run 1.** `scripts/preserve-session-records.sh` reads `.superpowers/sdd/tasks/progress.md` and
nothing else, so the original path would have been dropped at cleanup — the ledger would have
existed for the whole change and then vanished with the worktree, which is the one outcome the
preservation step exists to prevent.
