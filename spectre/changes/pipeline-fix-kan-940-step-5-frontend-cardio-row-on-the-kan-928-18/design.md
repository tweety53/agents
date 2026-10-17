## Context

- Evidence: gymie-playwright reflog for `spectre/kan-940-step-5-frontend-cardio-row-on-the-kan-928`
  opens `branch: Created from origin/main` at 2026-10-16 23:18:24 +0300; the session transcript
  (`~/.claude/projects/-Users-tweety53-Projects-gymie/3624d177-….jsonl`) shows the parent's command
  `git worktree add -q -b spectre/$N ../gymie-playwright-worktrees/$N origin/main` while its last
  stage mark was `flow.review-panel`; the same session had created the gymie-frontend worktree from
  `origin/feature/cardio` with the flowBase set. kan-940's `link.md` `## Merge order` names `.` and
  `gymie-frontend` only, so `flow.isolate-workspace` never created the playwright worktree.
- `kickoff-worktree.sh` already does every step the `## apps` recipe spells out — location guard,
  fetch, resume-or-cut, flowBase record, state persist, `## worktree setup`, push — and refuses a
  missing `origin/<base>` with exit 1, nothing added. Its one gap for a non-canonical repository is
  the state record: it runs `flow state add-worktree -C <project>`, and `fallback.ProjectKey` keys
  the record by that directory's main checkout, so a peer repository's add misses the change's
  STARTED record. A `-C` naming any worktree of the canonical repository keys it correctly.
- Default branches differ across a project's repositories (gymie `develop`, gymie-playwright
  `main`) on every ordinary change, so a guard on resolved bases alone would refuse every
  multi-repo landing; only a base a worktree records is the change's base.

## Decisions

### One script for every added worktree

**ID:** kickoff-from
**Status:** active
**Chosen:** `kickoff-worktree.sh --from <change worktree> <repo> <name> [<base>]` — the base
defaults to `git -C <change worktree> config --get branch.spectre/<name>.flowBase`, an explicit
`<base>` still winning, and `flow state add-worktree` runs with `-C <change worktree>`; the prose
recipe in `cross-repo-worktrees.md` and any later stage call it.
**Considered:** a new `add-change-worktree.sh` — a second copy of kickoff's six steps; tightening
the prose alone — the prose recipe was already right and the parent improvised past it, so the
fix has to make the right path one call; reading flowBase off a sibling worktree inside
`resolve-base-branch.sh` — resolution would then depend on which sibling worktrees are on disk.

### A missing base stops the run

**ID:** missing-base-stops
**Status:** active
**Chosen:** keep the existing contract — `cross-repo-worktrees.md` makes a failed worktree add a
hard failure that stops the run, and kickoff's missing-base exit 1 adds nothing; the message now
names the two answers the operator chooses between: push `<base>` to that repository's origin, or
re-run naming that repository's default branch as the explicit `<base>`.
**Considered:** falling back to the default branch silently — exactly the defect; asking a
question from inside the script — a script cannot ask, and the stop already reaches the operator.

### The integrate guard keys on a recorded base

**ID:** base-mismatch-guard
**Status:** active
**Chosen:** in `prepare`, after every worktree's base resolves and only on a run-1 path: for each
worktree whose branch records flowBase `<b>`, every other worktree resolving a base other than
`<b>` while `refs/remotes/origin/<b>` exists in its repository stops `STOP: base-mismatch`, before
any stage mark, rebase or push. The parent asks the operator (`skills/flow/integrate.md`).
**Considered:** refusing whenever resolved bases differ — refuses every ordinary gymie +
gymie-playwright change (`develop` vs `main`); an `--accept-base-mismatch` escape — no case needs
landing a worktree off the change's base when that base exists in its repository; the same check
in `commit` and `land` — both run after `prepare`, which a mismatched set can no longer pass.

### Memory follows the tree

**ID:** cardio-memory
**Status:** active
**Chosen:** update `cardio-feature-branch.md` and its `MEMORY.md` index line: `feature/cardio`
exists in gymie, gymie-frontend and gymie-playwright (`origin/feature/cardio` in gymie-playwright
since 2026-10-14), so every cardio worktree, playwright's included, is cut from and lands on it.
`quick-mode-feature-branch.md`'s "playwright: main" records a merge that happened and stays.
**Considered:** leaving it — it tells a later session the epic's playwright work goes elsewhere.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `medium` — Task 3 adds a stop to integrate's prepare that every multi-repo landing
crosses, and Task 1 changes which project's state record a worktree add writes; a wrong key or a
too-broad refusal passes a skim.
**Considered:** `low` — Task 2 alone is prose.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, g1: Tasks 1–3 in plan order, one fixer — Task 2 cites Task 1's flag, and
the three are small.
**Considered:** Task 3 as a second group — a second dispatch for one shell function and one case.

## Open questions

## Live check

None: the change edits a script, a guard and prose, and no running service holds state it alters;
`scripts/test-integrate-change.sh` and the Go tests exercise both mechanisms against real git
repositories.
