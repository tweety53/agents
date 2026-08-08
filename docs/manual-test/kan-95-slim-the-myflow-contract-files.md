# Manual test — kan-95-slim-the-myflow-contract-files

This repository declares **no runnable application**: it is the source of the myflow skills,
commands and rules, installed elsewhere by `setup.sh`. "Running the apps" here means running its
guard scripts, its assertion harnesses, and a sandboxed installer pass.

Run everything from the apply worktree:

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-95-slim-the-myflow-contract-files
```

## The guards and harnesses

- [ ] check `scripts/check-contract-budget.sh` — reports `BUDGET-OK: 24 contract file(s) within budget`
- [ ] check the count is right — 24 is 16 contracts plus 8 skill files; a wrong count is how a run that scanned the wrong tree would look
- [ ] check a covered file with no budget row fails — `mkdir -p skills/scratch && touch skills/scratch/SKILL.md`, re-run, confirm it names that file, then delete the directory
- [ ] check a file over budget fails — append 60KB to `skills/myflow-contracts/pipeline.md`, re-run, confirm it names that file, then `git checkout -- skills/myflow-contracts/pipeline.md`
- [ ] check it refuses a symlink — `ln -s /etc/hosts skills/myflow-contracts/ghost.md`, re-run, confirm it names it as a symlink and never prints that file's size, then delete it
- [ ] check every lint command in `.myflow/project.md`'s `## lint` list exits 0
- [ ] check every harness in `.myflow/project.md`'s `## test` list passes

## The per-run load, which is what this change bought

- [ ] check the figure — `wc -c skills/myflow-do/SKILL.md skills/myflow-contracts/{pipeline,project-configuration,workspace-isolation,state-file,jira-integration}.md` totals about 166,534 bytes, against 210,481 before
- [ ] check no appendix is loaded by a run — `grep -rn 'rationale\.md' skills/myflow-*/SKILL.md` returns only citations, never a load instruction
- [ ] check `README.md` is not loaded by a run either — `grep -rn '(`README\.md`)' skills/` returns nothing that a command reads to follow a rule

## `/myflow-info` is gone

- [ ] check both command trees — no `myflow-info.md` under `commands/` or `commands-claude/`
- [ ] check the skill directory is gone — no `skills/myflow-info/`
- [ ] check nothing still offers it — `grep -rn 'myflow-info' CLAUDE.md AGENTS.md README.md skills/README.md rules/` returns nothing
- [ ] check `README.md` carries what it used to explain — a `## How the pipeline works` section with the state diagram, a level-1 stage table, and a level-2 expansion for each of the eight stages
- [ ] check the surface is described as three pipeline commands plus one read-only one, everywhere it is described

## State self-heal is gone

- [ ] check the contract file is gone — no `skills/myflow-contracts/state-self-heal.md`
- [ ] check nothing cites it — `grep -rni 'state.self.heal' skills/ scripts/ rules/ commands/ commands-claude/ CLAUDE.md AGENTS.md README.md` returns nothing
- [ ] check `skills/myflow-status/SKILL.md` never writes a state file — it reports a missing or unparseable one and omits that change from the table
- [ ] check the unparseable definition survived — `skills/myflow-contracts/state-file.md` defines it directly rather than citing the deleted file

## `/myflow-status` is slimmer

- [ ] check the table has six columns — Change, Jira, State, PR, Next, Updated, and no worktree column
- [ ] check it makes no network call — `grep -n 'gh ' skills/myflow-status/SKILL.md` returns nothing
- [ ] check the detail view still gives the absolute worktree path
- [ ] check the PR column's limitation is stated — it shows the recorded pull request's number and not whether it is open

## Run the pipeline against a real change, which is the only end-to-end test

This is the part a guard cannot do. Install from this worktree into a sandbox, then exercise the
commands whose contracts changed.

- [ ] check the installer — `SANDBOX="$(mktemp -d)"; HOME="$SANDBOX" ./setup.sh global`, then confirm `$SANDBOX/.claude/skills/myflow-contracts/` holds every file the index names, including the three new ones
- [ ] check the installed tree has no `myflow-info` — neither a skill directory nor a command file
- [ ] check `/myflow-status` with no argument on a real project — the six-column table renders, and no `gh` process runs
- [ ] check `/myflow-status <name>` on a change at `STARTED` — the regenerated block carries the folded `**Recorded:**` line with decisions, open questions, effort and all three models
- [ ] check `/myflow-start` reaches its convergence prompts — the rule governing the stage's exit now lives in `pipeline.md`, so confirm the run can follow it without reading `README.md`

## Known incomplete

- The per-run figure is **166,534 bytes against the proposal's projected ~128,500**. The reduction is
  real — 210,481 → 166,534, or 20.88% — but the projection was not met and will not be by this change.
  Three of the four evicted files classified more conservatively than projected, each deliberately and
  each confirmed by review; and five fix rounds correctly put rules *back* into loaded files, which
  moved the figure the wrong way by design. Restoring a rule to where a run can read it beats a
  smaller number.
- **This change needed 4 review-panel passes and 5 fix rounds to reach zero open findings**, and a
  large share of the findings were defects introduced by earlier fixes rather than by the original
  ten tasks. Every round's signature failure was the same: change one location, leave a second
  description of the same thing standing somewhere else. Two were Critical and neither was detectable
  by any guard in this repository — a citation resolving into the target project's own tree, and a
  guard that would have blocked every first `/myflow-do` run. Worth weighing when judging whether the
  five parts should have shipped as one change.
- `scripts/check-references.sh` still cannot detect a hollowed section, and cannot tell whether a
  citation reaches the *right* content. That is KAN-84, still open. This change verified by reading
  plus a per-move ledger per task and per fix round; nothing mechanical covers it.
- **KAN-101 filed at Highest** — pre-existing, not this change: bare `scripts/*.sh` citations in
  installed files resolve into the target project, so a pull request adding
  `scripts/check-unfinished-work.sh` gets it **executed** by `/myflow-finish` run 1. Strictly worse
  than anything this change introduced, and deliberately not folded into it.
- One pre-existing spec inaccuracy noted and left alone: `myflow-contract-economy`'s scenario
  "Follow-up issues moves to its own file" asserts `jira-integration.md` "does not contain" that
  section, while a redirect stub carrying the heading does exist there. Predates this branch.
