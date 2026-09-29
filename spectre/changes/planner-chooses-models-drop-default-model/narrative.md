# planner-chooses-models-drop-default-model — session narrative

## 2026-09-29 — creating run

- **Resumed at `STARTED`, inline execution.** This run's own model resolution was the last to read
  a `defaultModel`: the store answered, so the recorded decision carries `opus`/`store` — the
  mechanism this change removes.
- **Build-red intermediate tasks.** Tasks 1 and 2 removed symbols (`DefaultModel`,
  `ErrInvalidModel`) whose last callers only Task 3 deleted, so the tree was red between them; the
  plan's Correction paragraphs record it. `mkSpace` lived in the deleted `modelkeys.go` and moved to
  `workspaceisolation.go` as `wiSpace`.
- **Guard hits along the way.** `check-references` (rewraps), `check-installed-citations` (a
  redundant line in `archive.md`), `check-self-review-report` (the angle table moved into
  `skills/flow-self-review/SKILL.md`; the guard was repointed), `check-plan-shape` (a test path under
  `Tests: none`, a Correction line opening with `**Files:**`, a missing `After:`), and
  `check-plan-provenance` (the 51-test count needed its `measured:` comment).
- **Operator decision — fixer pair.** Inline runs have no panel-fix dispatch, so the recorded
  `inline-fixer-planner-chosen` decision was false; the operator chose to supersede it with
  `fixer-pair-sdd-only`.
- **Operator decision — base moved.** `origin/main` moved 2 commits with no overlap at panel round
  1; the operator chose to continue without a rebase. Integrate's sync rebases.
- **Panel.** Round 0 raised F1 (Important: the gated reviewer could run on sonnet) and F2–F6
  (Minor). All fixed. F6's first fix left two citations at the forwarding pointer; the principles
  re-run caught it, fixed in round 2. F5's mutation first survived a `Contains` assertion; the test
  was pinned with `HasSuffix`.
- **Environment.** zsh does not word-split unquoted variables, which broke the first pathspec
  commit; `run-reproducer.sh` takes its flags after the positionals and pins the reproducer file's
  sha256, and a reproducer needs its execute bit. A `flow stage begin` issued from the skills
  directory left `flow.verify`'s run unmatched by its `end`; the next stage's `begin` superseded it.
- **Gated-review fixer key.** `check-panel-fix-single-dispatch.sh` flagged the gated per-task fix
  row key `task-1+4+5-implementer-fix-1` as out of the panel-fix shape; it is not a panel-fix
  dispatch, and the prompt was auto-resolved on Continue.

## 2026-09-29 — integrate run

- Preflight: `STAGED-CLEAN`, `DRIFT-CLEAN`, `RUN1` against `origin/main`.
- Unfinished-work gate: `CLEAR`; visual verify `VISUAL-VERIFY-OK` (no UI paths).
- Base: `check-base-moved.sh` `CLEAR` — no rebase needed.
- Route: merge and push, from the project's configured default, not asked.
- Environment: one `flow stage end` issued from `~/.claude/skills` failed to resolve the project key
  (not a git checkout); re-issued from the main checkout.
