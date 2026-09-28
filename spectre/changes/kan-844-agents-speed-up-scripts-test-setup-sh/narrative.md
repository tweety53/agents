# kan-844-agents-speed-up-scripts-test-setup-sh — session narrative

## 2026-09-28 — creating run

- **Base moved mid-run.** kan-761 landed on main with an overlap on `skills/flow/implement.md`. The
  operator chose to rebase onto main (`979ee6cb`), and the branch was force-pushed with lease; the
  task commits' shas changed with it.
- **Incident — a real `setup.sh global` against the operator's HOME.** A perl `s|..|..|` with
  escaped pipes mangled `scripts/test-setup-agents.sh`, and running it executed `./setup.sh global`
  unsandboxed at 12:21:48, relinking `~/.claude` and `~/.zcode` into this worktree. The script was
  restored from git and re-edited with Python; with the operator's approval `setup.sh global` was
  re-run from the main checkout, and no link points into the worktree afterwards. Recorded with
  `flow record incident`. Lessons: never use perl `s|||` on shell text, and always `git diff` a
  script before executing it.
- **The operator's `~/.cache/flow-guard` was written at 12:22** by pre-fix runs of the ported
  harness (findings F1/F5/F8), before the guard run got its own `FLOW_GUARD_CACHE_DIR`. Those cache
  entries are harmless build products but are the operator's to clear.
- **Contract friction — PLAN FIELDS vs the record guards.** Correcting a shipped task's
  `**Files:**`/`**Tests:**`/`**Baseline:**` fields after the fix round broke
  `check-task-records.sh` and plan provenance, so the fields were restored to what each task's own
  commit carries and the post-panel additions are stated in dated Correction paragraphs instead.
- **`git add` with an `:(exclude)` pathspec staged nothing for new files** after the clearing
  reset; files were staged plainly.
- **Reproducers vs a refactoring fix.** Seven round-0 reproducers pinned their premises to lines
  the fix moved, or mutated code the fix refactored, so `run-reproducer.sh` refused them as
  ambiguous. They were re-authored with file-level premises and proved both ways with
  `prove-reproducer.sh`. Mutation reproducers `git archive HEAD`, so the fix had to be committed
  before they could flip.
- **The SIGINT proof was vacuous at first**: a background job ignores SIGINT, so the mutant and the
  fix behaved alike. The proof was redone with SIGTERM (fixed: exit 130, no sandbox left; mutant:
  exit 143, sandbox left).
- **Operator decision F4**: the gated reviewer of an `xhigh` implementer group runs at `high`.
- **Verify blocked on another change.** `check-task-records.sh` failed on
  `withdraw-changes-abandoned-before-planning`, which landed on main as one squash commit
  (`a0cc01e7`) whose archive never reached main; it failed identically on main. The operator chose
  to fix it on this branch: that plan's `**Commit:**` subjects now name the squash commit, with a
  dated Correction note (`a5e7e207`).
- **Deferred:** F20 (Minor, interrupt-path cleanup race) went to `KNOWN-BUGS.md`, beside the two
  task-review Minors.
