# kan-1073-commit-only-what-is-needed-permanently

## Why

Every `/flow` run commits material it needs only while it runs: the change directory
(`spectre/changes/<name>/` — proposal, design, plan, narratives, link records, visual evidence) in
`chore(spectre): plan` and `chore(spectre): merge order` commits, a `chore(spectre): drop` at
landing, and a `docs(self-review)` report afterwards. The KAN-1072 audit
(`docs/audits/2026-10-31-spectre-archive-audit.md`) found these commits make up a large share of
the landed history of six repositories, that the drop leaves residue behind (agents `50e0a1f7`
cleaned the last of it), and that a dozen guards and a whole package (`stats/internal/landed`)
exist only to find the plan again in history once the drop removed it. The umbrella follow-up the
audit names (items 1–4) removes the committed visual evidence, the committed self-review report,
the narrative-at-drop lesson source and the spectre `archive`/`refs` commands one by one; this
change removes their common cause instead.

## What changes

- `spectre/changes/` is gitignored in agents, spectre, gymie, gymie-frontend, gymie-playwright and
  flow-eval. A change's directory lives only on disk in its worktree, exactly where it lives today,
  and is never staged, committed, dropped, archived or restored.
- No run writes a planning commit, a drop commit, a self-review report or a self-review report
  commit. A landed branch carries the change's product commits only.
- Every guard that read the plan from a commit or from history reads the on-disk file in the
  worktree. `stats/internal/landed` and every live-else-archive fallback are deleted.
- flowd keeps the two files that outlive the worktree, in a new `change_files` table:
  `tasks.md` (the board's plan of a finished change) and `self-review-context.md` (a deferred
  `/flow-self-review`'s whole input). `flow change save <name>` writes them once, in run 2's
  cleanup, and `remove-change-worktrees` refuses until it ran; `flow change file <name> <path>`
  prints one.
- The self-review report is gone: its findings already live in flowd, and the land check reads the
  `flow.self-review` stage row instead of a committed file.
- The rejection search over archived designs (kan-970) is deleted.
- spectre loses `spectre archive`, `spectre refs`, cross-tree requirement citations and every
  archive-aware code path; its vendored planning skill is re-copied from agents.
- The guards, scripts, stages (`flow.commit-archive`, `flow.sync-archive`) and contract text that
  existed only for planning commits, the drop and the report are deleted.
- Where a planning commit used to be made — the plan gate, `/flow-plan`'s capture and a fix run's
  tasks append — the run copies `tasks.md`, `proposal.md` and `design.md` into a numbered snapshot
  `.superpowers/sdd/<name>/plan-snapshots/<k>/`, never committed.
- flow-eval's `case add impl|review` reads the change's worktree on disk, never git history: for
  task n the earliest plan snapshot holding it, so no later review-fix Correction leaks the answer
  into the case, and the ledger and panel record only as the hidden grading answer. It runs in the
  self-review inside integrate run 1 before cleanup; once the worktree is gone it refuses with
  `plan files gone — mine before cleanup`.
- Each repository's tip loses every file still tracked under `spectre/changes/` and
  `docs/self-review/`, and the placeholder `spectre/specs/` where nothing else lives there. History
  keeps every file.
