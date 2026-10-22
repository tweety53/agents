## Context

The four findings are kan-1006's self-review pass, planned as one change per
`skills/flow-self-review/SKILL.md` step 3. Each was checked against history before its fix was
designed: no commit since 2026-10-21 on `renderslotprompt.go`, `protect-main-checkout.py` or
`review-panel.md` addresses either F-A or F-C, and no archived change records a rejected
REPRODUCER paragraph or restore denial. F-D's rows were read live: `flow record dispatches -change
kan-1006-list-mode-step-1-backend-workout-mode-preference` shows rows 6966 and 6967 with no
`agentId` and `"metrics":{}`, and both transcripts exist.

Every prose and code edit was applied and checked in a scratch clone of this branch at `0e0d8a42`
(gymie's at `894cd84c`): the diffs are this change's `patches/task-<n>.diff`, and the sentences the
prose adds are acknowledged in `verbatim-moves.txt` beside this file.

## Live check

- **Hook (Task 3).** After the landing reinstalls the hook, in any linked worktree of this
  repository: make an unstaged edit to a tracked file and pipe a PreToolUse payload for
  `git -C <worktree> checkout -- <file>` into `python3 ~/.claude/hooks/protect-main-checkout.py`.
  Before: no output (allowed). After: a `deny` naming the file and `break-and-prove.sh`. No output
  after the landing means the installed hook is still the old one.
- **Store (Task 4).** `flow record dispatches -C /Users/tweety53/Projects/gymie -change
  kan-1006-list-mode-step-1-backend-workout-mode-preference` — before: rows 6966 and 6967 carry no
  `agentId` and `"metrics":{}`; after: `a62f95e81cc0b07c5` and `a01308edf624bb805` respectively, with
  non-zero sidechain token figures. A missing agent id or empty metrics means the repair never
  reached the store.
- **Slot prompt (Task 1).** The next panel round's rendered `slot-prompt-<round>-<slot>.md` carries
  the `**REPRODUCER:**` block once; its absence means the run used an old `flow-guard` build.

## Decisions

### The REPRODUCER paragraph is a dispatch block, rendered and paragraph-checked

**ID:** reproducer-dispatch-block
**Status:** active
**Chosen:** a verbatim `> **REPRODUCER:**` block in `skills/flow/review-panel.md` beside REPRODUCE,
DON'T READ; `render-slot-prompt` adds `REPRODUCER` to its shared-label list and
`check-dispatch-paragraphs` requires it at review-panel.md (min 1) with the phrases "exits non-zero
when the defect is present", "`# demonstrates: <path>:<line>:<content>`", "never a prose
description" and "|| true" — the same mechanism every other always-carried slot paragraph uses.
**Considered:** pasting the rule into each reviewer template — five copies that drift; relying on
the renderer's own exit 2 when the block is missing — catches a deleted block but not a dropped
phrase.

### The block carries no renderer placeholder

**ID:** reproducer-block-no-placeholder
**Status:** active
**Chosen:** the block names the directory as the canonical worktree's
`<project>/.superpowers/sdd/reproducers/`, so it renders unchanged.
**Considered:** `<abs-worktree>/…` — `rspUnfilled` refuses it unless the renderer substitutes it, a
second code path for one path; `<round>-<id>-<n>.sh` likewise.

### The REPRODUCER block restates the rule above it

**ID:** reproducer-block-restates
**Status:** active
**Chosen:** the block condenses the reproducer rule the paragraphs above it state, as the
other dispatch blocks condense theirs: a slot reads its prompt, never review-panel.md, and the
missing carry is the defect.
**Considered:** cutting the long-form rule to the block alone — loses the guard and runner contract
the parent reads.

### The before backend is stopped by its port listener

**ID:** before-backend-port-kill
**Status:** active
**Chosen:** `lsof -ti tcp:<free port> -sTCP:LISTEN | xargs kill` in gymie's recipe, naming why
`./gradlew --stop` is wrong. The edit rides kan-1006's change branch in the gymie worktree with a
`Flow-Key:` trailer.
**Considered:** `./gradlew --stop` from the before worktree — stops every daemon of that Gradle
version, the defect; a `devBootRunStop` Gradle task — new tooling for a one-line recipe.

### The restore denial extends protect-main-checkout.py

**ID:** restore-denial-in-protect-hook
**Status:** active
**Chosen:** a `discards_unstaged` check in `hooks/protect-main-checkout.py`'s git-verb scan, scoped
to linked worktrees, judging `git checkout … -- <path>` and `git restore <path>` (not
`--staged` alone) by `git diff --name-only -- <paths>`; it fails open like the rest of the hook.
A `git checkout <path>` without `--` is not judged (a `ponytail:` note names it).
**Considered:** a new hook file — a second PreToolUse registration and its own command-text
tokenizer, which this hook already has; judging every repository — the main checkout on its
default branch already denies `restore`, and the finding is about flow worktrees.

### The restore rule's canonical home is git-boundaries.md

**ID:** restore-rule-home
**Status:** active
**Chosen:** a `## Restoring a path with unstaged edits` section in
`skills/flow-contracts/git-boundaries.md`, cited by one sentence after MUTATION PROOF in
review-panel.md; the MUTATION PROOF block itself is unchanged, so its paragraph check stands.
**Considered:** a sentence inside the MUTATION PROOF block — the inline parent never reads that
block, and the hook's reason needs a contract heading to cite.

### Patches live beside the plan

**ID:** patches-beside-plan
**Status:** active
**Chosen:** each code-and-prose task applies `patches/task-<n>.diff`, proved in a scratch clone, so
prose lands exactly as `verbatim-moves.txt` acknowledges it.
**Considered:** prose described in steps — the fixer's rewording fails `check-verbatim-moves.sh`.

### The store repair commits nothing

**ID:** store-repair-no-commit
**Status:** active
**Chosen:** Task 4 runs `flowd restamp-dispatch` for each row and verifies with `flow record
dispatches`; `**Commit:** none`.
**Considered:** folding it into a code task — no code changes, 90efd086 already fixed the race.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — the change adds a PreToolUse denial to a hook every session runs and a
new required dispatch paragraph to a guard, beyond prose.
**Considered:** flow-low — fits prose-only changes, not a new hook verdict.

### Fixer groups

**ID:** fixer-groups
**Status:** active
**Chosen:** one fixer dispatch, g1, Tasks 1, 2, 3, 4 in series — Tasks 1 and 3 share
`skills/flow/review-panel.md` and must apply in order, Task 2 is a one-patch gymie commit, and
Task 4 is two store commands.
**Considered:** a dispatch per repository — three cold starts for four patch-sized tasks.

## Open questions

- **Correction (review-1):** `restore-denial-in-protect-hook` widened — when `checkout` names a tree-ish before `--`, or `restore` takes `-s`/`--source`, the dirty check diffs against HEAD so staged-only edits are protected too; redirect tokens end the path scan.
