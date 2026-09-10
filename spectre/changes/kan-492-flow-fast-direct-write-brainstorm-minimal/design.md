## Context

Explored via `/flow-plan` against `skills/flow-fast/*.md` (kan-490), the contracts it cites, the
dev stats store and the KAN Jira workflow. The operator named brainstorm+artifacts and finish as
the slow phases; the `flow` Go-app interaction (stage marks, records, state) stays in full. Full
research: `docs/superpowers/research/kan-492.md` (deleted once adopted, per task 1 below).

## Decisions

- **Brainstorm — direct write.** Read context, one batched `AskUserQuestion` only for true
  blockers, then write the three artifacts straight into `spectre/changes/<name>/`. No
  `superpowers:brainstorming`, no `superpowers:writing-plans`, no `docs/superpowers/specs/` file.
  Supersedes kan-490's brainstorm decision for `/flow-fast`.
- **Artifacts — minimal shape.** `tasks.md`: checkbox + `**Files:**`/`**Tests:**`/`**Commit:**`
  only — no `**Regression:**`/`**Baseline:**`/`**Squash-with:**`/`**After:**`, no
  `**Execution:**`/`**Relocation:**` headers, no plan-provenance or build-green tag, no
  `check-plan-shape.sh`. `design.md`: short prose `## Context`/`## Decisions`, no `ID:`/`Status:`
  entries. No `flow record render` and no `docs/superpowers/` commit at finish — the store is the
  terminal record. Consequence accepted: a `/flow` resume of a `/flow-fast` change needs plan
  repair before its guards pass. Supersedes kan-490's artifact-shape decision.
- **Finish — one merge, one push, no run 2.** Merge-and-push only; the landing question is never
  asked. Guards (`check-finish-preflight.sh`, `check-unfinished-work.sh`, `check-base-moved.sh`) →
  `git reset --soft` + `commit-split.sh` → one landing worktree on `<base>` → `git merge --no-ff`
  → `spectre archive` on `<base>` → `check-archive-scope.sh` → commit → one `git push origin
  <base>` → cleanup → `FINISHED` → Jira Done. No `chore/archive-<name>` branch, no second
  merge/push, no remote branch delete. Supersedes kan-490's finish decision.
- **Cleanup — stop + process scan only.** Checks 1–4 (incl. the ignored-files disclosure) are
  skipped — checks 1–3 are true by construction after commit-two + merge in the same invocation,
  and check 4's disclosure is in the store. Checks 5 (`## stop`) and 6
  (`check-worktree-processes.sh`) stay as gates. Supersedes kan-490's guard-set decision (cleanup
  half).
- **Jira — straight to Done.** `In Progress` at kickoff (unchanged); one `Done` transition after
  the push, no `In Review` hop — KAN's workflow offers `Done` directly from `In Progress`.
- **Review panel — shared dispatch file, narrower re-run.** The three dispatch paragraphs and the
  `final-review.diff` step move into `skills/flow/panel-dispatch.md`, cited by both
  `review-panel.md` and `skills/flow-fast/review.md`. After a Critical/Major fix, only
  `simple-reviewer` re-runs on the delta — `primary`'s pass-1 plan-alignment verdict stands.
  Supersedes kan-490's guard-set decision (review half).
- **Handoff — literal block.** `skills/flow-fast/review.md` carries its own `IN_PROGRESS` block
  and run instructions verbatim instead of citing `handoff-blocks.md`.
- **Stage marks — unchanged, chained.** All twenty `flow.*` keys stay; each stage's `end` mark
  rides in the same call as the next stage's `begin` plus its first work.
- **Spec fixes folded in.** `brainstorm.md` section D's `check-plan-shape.sh` sentence goes (the
  guard-set contradiction resolved in the guard set's favour); the `/myflow-fast` names in
  `finish-contract-run1.md`/`finish-contract-run2.md` become `/flow-fast` or are deleted where the
  step no longer applies.

Rejected alternatives, and why, are recorded per-decision in
`docs/superpowers/research/kan-492.md` §2 before its deletion; the reasons are not restated here a
second time since the note is this change's own record until task 1 removes it.
