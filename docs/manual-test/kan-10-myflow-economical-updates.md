# Manual test — kan-10-myflow-economical-updates

**Purpose:** Cut myflow's fixed per-session token cost — move four contract sections out of the always-on rule file into an on-demand skill, add a script for the mechanical state write, and move one review-panel slot to the economy tier.
**Branch / worktree:** `openspec/kan-10-myflow-economical-updates` · myflow sources `/Users/tweety53/Projects/agents-worktrees/openspec-kan-10-myflow-economical-updates`
**Next after sign-off:** `/myflow-review kan-10-myflow-economical-updates`
**Manual test status:** SKIPPED — 2026-07-28 (Gate C intentionally bypassed)

> Every box below is deliberately left unchecked. Nothing here was run as part of Gate C. The
> checklist is retained so it can be worked through later if the bypass is reconsidered, and so
> `/myflow-review` can report what was not verified rather than implying it was.

## How to run (involved apps)

This repository has **no runnable application**. It is the source of the myflow skills, commands
and rules, which `setup.sh` installs into a home directory. There is nothing to start, no port and
no URL, so "running it" means installing it into a throwaway `HOME` and inspecting the result.

**Prerequisites.** Run every command from the apply worktree below — not from
`/Users/tweety53/Projects/agents`, which is the main checkout on `main` and does not contain this
change.

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-10-myflow-economical-updates
```

### Install into a sandboxed HOME

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-10-myflow-economical-updates
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
echo "$SANDBOX"
```

Nothing outside `$SANDBOX` is written. Your real `~/.claude`, `~/.cursor` and `~/.codex` are
untouched — `scripts/test-setup.sh` asserts this, but the sandbox makes it true by construction.

### Verification scripts

```bash
cd /Users/tweety53/Projects/agents-worktrees/openspec-kan-10-myflow-economical-updates
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-state-advance.sh
```

## Functionality checklist

### The always-on rule layer shrank (myflow-contract-distribution)

- [ ] `wc -c rules/myflow-manual-review.mdc` reports **≤ 32768 bytes** (was 58163 at `ee7d143`)
- [ ] The rule file contains no state-write `jq` template, no `## standards` entry-form table, no containment rules, and no Jira transition table
- [ ] The Pipeline stages, Stage transitions and Stage boundaries tables are all still present in full
- [ ] `skills/myflow-contracts/` contains exactly `SKILL.md`, `state-file.md`, `state-self-heal.md`, `project-configuration.md`, `jira-integration.md`
- [ ] Each of the four vacated sections still has its `##` heading in the rule file, followed by a stub naming what the contract governs **and** the file that now holds it
- [ ] Opening `skills/myflow-contracts/jira-integration.md` alone is enough to follow the Jira contract — it reads as a standalone document
- [ ] In the sandbox: `myflow-contracts` is present under `$SANDBOX/.claude/skills/`, `$SANDBOX/.cursor/skills/` and `$SANDBOX/.codex/skills/`
- [ ] In the sandbox: the managed block in `$SANDBOX/.claude/CLAUDE.md` does **not** contain the contract bodies, and the file is roughly 26 KB smaller than before this change
- [ ] In the sandbox: `grep -c 'myflow:begin' "$SANDBOX/.claude/CLAUDE.md"` returns exactly `1`

### The state write is mechanical (myflow-state-advance)

- [ ] A happy-path advance writes the new `stage`, a fresh `updatedAt`, and the invoking command in `updatedBy`
- [ ] Every field the script does not own survives byte-identically — all four gates, `artifactUrl`, `jiraIssue`, `fastPath`, `REVIEWED_TREE`, `MERGE_BASE`
- [ ] A stage outside `--accepted` exits **4** and leaves the state file byte-identical
- [ ] A missing or non-object state file exits **3** and creates nothing
- [ ] A `worktree` path git no longer lists exits **3** — and a valid worktree belonging to *another* repo does **not** (run the script from a different repository to confirm)
- [ ] `--target originStage` with `do-review-started` resolves to `awaiting-do-review` and clears `originStage`
- [ ] An `originStage` outside the six legal origins exits **6** and is never repaired into a plausible stage
- [ ] A value-taking flag given as the final argument exits **2** with a usage message on stderr
- [ ] `--name '../../../../elsewhere'` is rejected with exit **2** rather than writing outside the state directory
- [ ] No Jira call is made for a change whose `jiraIssue` is set

### Commands use the script (myflow-state-advance, requirements 4-5)

- [ ] All seven `*-done` / `*-manual-review` files in **both** `commands/` and `commands-claude/` instruct the agent to run `state-advance.sh` before loading the skill
- [ ] Each file's `--target` / `--accepted` values match its own `TARGET_STAGE` / `ACCEPTED_STAGES` and the Stage transitions table
- [ ] `commands-claude/` uses `~/.claude/skills/…` and `commands/` uses `~/.cursor/skills/…`
- [ ] Every non-zero exit routes to the skill — including 127, the partial-install case
- [ ] `skills/myflow-contracts/state-self-heal.md` states which checks the script performs, which it skips, and why that is acceptable

### Review-panel economics (myflow-review-panel-economics)

- [ ] The panel table shows slot 4 (Adversarial) on the **economy** tier with `model` required
- [ ] Slots 0 (Primary) and 2 (Principles) still inherit the parent model with no override
- [ ] `grep -rn "slots 5+ only" skills/` returns nothing
- [ ] `rules/myflow-manual-review.mdc` names slot 4 as economy-tier — it is read before the skill, so a stale statement there would silently override it
- [ ] The fast path still dispatches only its three required slots; slot 4 is not among them

### This repo verifies itself (agents-repo-verification)

- [ ] Every command in `.myflow/project.md`'s `## test` and `## lint` blocks runs from the repo root and exits 0, exactly as written
- [ ] `## lint` states explicitly that no auto-fix command exists in this repository
- [ ] `## apps` states there is no runnable application and names the guard scripts as the verification surface
- [ ] `scripts/check-references.sh` takes no arguments and scans the same set from any directory in the repo (try it from `/tmp` and from `skills/`)
- [ ] Renaming a referenced section makes the guard fail with `file:line`; restoring it makes the guard pass
- [ ] A reference resolving outside the repository root fails, identically whether or not that file exists
- [ ] `skills/openspec-review-superpowers/SKILL.md` runs all five scripts in its verification gate
- [ ] No `refs-guard:allow` suppression markers exist anywhere in the scanned tree

## Sign-off

- [ ] All involved apps ran from their apply worktrees
- [ ] Every checklist item above verified
- [ ] Ready for `/myflow-review kan-10-myflow-economical-updates`
