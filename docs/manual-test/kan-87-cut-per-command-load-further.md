# Manual test — kan-87-cut-per-command-load-further

This repository declares **no runnable application**: it is the source of the myflow skills,
commands and rules, installed elsewhere by `setup.sh`. "Running the apps" here means running its
guard scripts, its assertion harnesses, and a sandboxed installer pass.

Run everything from the apply worktree:

```bash
cd /Users/tweety53/Projects/agents-worktrees/kan-87-cut-per-command-load-further
```

## The budget guard, at its widened scope

- [ ] check `scripts/check-contract-budget.sh` — reports `BUDGET-OK: 22 contract file(s) within budget`
- [ ] check it covers the skill files — append 60KB to `skills/myflow-do/SKILL.md`, re-run, confirm it names that file, then `git checkout --` it
- [ ] check it covers the appendices — append 20KB to `skills/myflow-do/SKILL-rationale.md`, re-run, confirm it names that file, then restore it
- [ ] check a new skill file with no budget row fails — `mkdir -p skills/scratch && touch skills/scratch/SKILL.md`, re-run, confirm it names it, then delete the directory
- [ ] check the count is right — the verdict must say 22, which is 13 contracts plus 9 skill files; a wrong count is how a run that scanned nothing would look
- [ ] check it refuses a symlink — `ln -s /etc/hosts skills/myflow-contracts/ghost.md`, re-run, confirm it names it as a symlink and never prints that file's size, then delete it
- [ ] check it is self-scoped — run it by absolute path from `/tmp`, confirm the same count rather than a clean run over nothing
- [ ] check `scripts/test-check-contract-budget.sh` — prints `all checks passed`

## The other guards and harnesses

- [ ] check `scripts/check-references.sh` — every citation across the corpus resolves
- [ ] check `scripts/check-vocabulary.sh`, `check-plan-provenance.sh`, `check-task-build-green.sh`, `check-workspace-isolation.sh` — all exit 0
- [ ] check every harness in `.myflow/project.md`'s `## test` list still passes

## The split itself

- [ ] check `pipeline.md` is 64,621 bytes and `finish-contract.md` holds the finish contract
- [ ] check nothing was lost — for each of the four split pairs, the line-multiset diff against `196f277` shows only removed lines that are a repointed citation, a deleted position word, or a rewrap continuation
- [ ] check each `SKILL.md` and its `SKILL-rationale.md` carry identical heading trees, fence-aware
- [ ] check no rule moved into an appendix — read the three `SKILL-rationale.md` files and `pipeline-rationale.md`, confirm nothing in them is something an agent must obey
- [ ] check `finish-contract.md` carries its rules **and** its reasoning inline, and that no `finish-contract-rationale.md` exists

## What each command now loads

- [ ] check `/myflow-info` still explains the pipeline from `pipeline.md` alone — the state diagram, the level-1 stage table and every level-2 finish expansion are in the core
- [ ] check `/myflow-info` tells a reader where the finish procedures went, and says plainly that it does not read that file
- [ ] check only `/myflow-finish` is instructed to load `finish-contract.md` — re-run `grep -ln 'finish-contract\.md' skills/myflow-*/SKILL.md` and read every hit rather than counting them
- [ ] check `skills/myflow-contracts/SKILL.md` names every file in that directory, and the directory holds every file it names
- [ ] check no skill instructs a `-rationale.md` load

## Distribution

- [ ] check a sandboxed `HOME="$(mktemp -d)" ./setup.sh global` places `finish-contract.md` and all three `SKILL-rationale.md` files under `.claude/skills/`, `.cursor/skills/` and `.codex/skills/`
- [ ] check `git diff -- setup.sh` is empty

## Run the commands against the split contracts

- [ ] check `/myflow-info` prints the pipeline explanation
- [ ] check `/myflow-status` renders correctly

## Known incomplete

- **The skill splits under-delivered, and the cause is the same one KAN-82 hit.** `myflow-do`
  40,529 → 37,885, `myflow-start` 28,944 → 25,999, `myflow-finish` 28,086 → 25,604 — about 9%
  against a 97,559-byte starting point. These files are rule-dense, and the
  `paragraph-granularity-partition` rule keeps any paragraph mixing a rule with its justification in
  the core. `jira-integration.md` behaved identically last change. `pipeline.md` is where the real
  win is: 88,253 → 64,621, a 27% cut.
- **`project-configuration.md` (45,258) and `workspace-isolation.md` (31,079) are still unsplit**,
  deliberately out of scope. Both are now ratcheted where they stand by the widened guard.
- **The `[ ! -r "$path" ]` readability pre-test has no independent test coverage.** Lens C found that
  removing it does not turn the harness red, because the `wc -c` guard behind it exits 2 for the same
  input. The exit-code contract still holds; the line is simply not independently proven.
- **A hand-authored table of `grep` hits went stale twice in this change** — corrected to four, then
  made five by a later fix in the same round. Both `tasks.md` and `design.md` now carry the rule to
  re-run the command as the task's last action rather than assert a count. Nothing enforces that
  rule; it is stated, like the corpus's other judgment rules.
- **Two reviewers wrote to the worktree during review**, one mutating the guard and one running
  `git checkout --` over three live unstaged edits. Both restored and disclosed, both verified by
  hand. Nothing is lost, but reviewers ran concurrently with edits in the same worktree, which is
  what made it possible.
