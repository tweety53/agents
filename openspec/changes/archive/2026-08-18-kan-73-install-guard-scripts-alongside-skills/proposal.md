# Install the myflow guard scripts alongside the skills

## Why

`setup.sh`'s `install_skills()` symlinks each `skills/<name>/` directory into
`~/.claude/skills/`, `~/.cursor/skills/` and `~/.codex/skills/`. Every guard script lives at
repo-root `scripts/` — outside that tree — and every skill and contract cites it as a bare,
repo-relative `scripts/<name>.sh`.

That path resolves only when the project a `/myflow-*` command is running against happens to
be this repository. In every other project it names nothing. Confirmed on this machine: no
`scripts/` directory exists under any installed `myflow-*` skill, and
`skills/myflow-finish/` carries none at all.

The contracts do define a hand-run fallback — "when the script is absent, perform the same
signals by hand" — so a hand-run is legitimate. The failure is that absence is discovered
**silently, at each call site**, and several checks specify behaviour prose cannot faithfully
reproduce: distinguishing a missing verdict line from a verdict, and checking the exit code
**as well as** the line.

Two incidents establish the cost:

- **KAN-70.** At `/myflow-finish` run 2, signal 1 of the preflight — `HEAD` against the
  recorded merge base — established that the branch carried no commits of its own and that
  everything was still staged. The ancestor test alone reports *merged* for such a branch,
  because a branch with no commits is an ancestor of every branch. Had the run trusted the
  ancestor test, run 2 would have archived the change and `--force`-removed both worktrees,
  holding roughly 12,000 lines of uncommitted work at that moment.
  <!-- measured: reported in KAN-73's own description, which records the figure from the KAN-70 run; not re-measurable here -->
- **KAN-37.** Both finish runs performed every check by hand. The hand-run cleanup
  verification caught a real leftover — an IDE-recreated `.idea/workspace.xml` surviving
  `git worktree remove --force`, which exited 0.

## What Changes

- **Guards ship with the skills.** Each command skill that invokes a guard gains a
  `scripts/` directory of git-tracked **relative symlinks** into repo-root `scripts/`, where
  the one real copy of each guard stays. The existing `install_skills()` symlink carries them
  with no installer change.
- **One resolution rule, stated once.** `pipeline.md` gains a section stating that a named
  guard resolves to `<the running command's own skill directory>/scripts/<name>`. Skills and
  contracts name guards by basename; no invoking call site carries a repo-relative
  `scripts/…` path any more.
- **Absence is reported once, loudly, at the start of a run** — one block naming every
  missing guard and the command that installs them — rather than silently at each call site.
  The run then proceeds under the hand-run fallback the contracts already define.
- **Two repo-root assumptions are fixed.** `plan-dispatch-bundles.sh` and
  `check-workspace-isolation.sh` derive a repository root as exactly one level above their own
  directory, which becomes a *skill* directory once they are reachable through a skill's
  `scripts/`.
- **A new guard, `check-guard-symlinks.sh`**, keeps all of the above true.

**Not in scope:** the panel-record marker linter the issue also names. It already exists —
`scripts/check-unfinished-work.sh` enforces every marker-block rule the issue lists, sharing
`scripts/lib/panel-record.sh` with `scripts/check-panel-reproducers.sh`. Building a second
guard over the same rules would recreate exactly the drift that library was extracted to
stop. See `design.md`.

**Not in scope:** any change to what a guard checks. This change moves guards and fixes how
they are addressed; it alters no check's logic.

**Carried, at the operator's instruction:** already-authored prose that was sitting uncommitted in
the main checkout when this change was planned — the `/myflow-fast` check-4 worktree-cleanup
override, stated in `skills/myflow-fast/SKILL.md` and named from
`skills/myflow-contracts/finish-contract.md` so the two files cannot silently disagree about a step
that destroys files. It is unrelated to KAN-73 in subject, and is carried here because it otherwise
has no change to land through and currently leaves `scripts/check-contract-budget.sh` failing in the
main checkout. It lands as its own task, in its own commit, touching no file this change would not
already touch.

The carried diff is 21 insertions across those two files, and nothing else.
<!-- measured: git diff --stat in the main checkout on 2026-08-17 — 2 files changed, 21 insertions(+) -->

## Impact

- **Affected specs:** `myflow-contract-distribution`
- **Affected code:** `skills/myflow-do/`, `skills/myflow-finish/`, `skills/myflow-status/`,
  `skills/myflow-fast/`, `skills/myflow-contracts/pipeline.md`,
  `skills/myflow-contracts/finish-contract.md`, `scripts/plan-dispatch-bundles.sh`,
  `scripts/check-workspace-isolation.sh`, `scripts/test-setup.sh`, `.myflow/project.md`
- **New files:** `scripts/check-guard-symlinks.sh`,
  `scripts/test-check-guard-symlinks.sh`, and the per-skill `scripts/` symlink directories
