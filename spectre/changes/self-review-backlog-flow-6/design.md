# self-review-backlog-flow-6

## Context

- Five Backlog Flow findings — the oldest five, KAN-702, KAN-714, KAN-715, KAN-718, KAN-719 —
  planned as one change per item 1 of **3. Fix every finding that is not `big`**
  (`skills/flow-self-review/SKILL.md`). Brainstorm design:
  `.superpowers/sdd/2026-10-18-self-review-backlog-flow-6-design.md` in this worktree. Every
  question was auto-resolved to its recommended option (`skills/flow-contracts/operator-prompts-auto-resolution.md`).
- **KAN-702 — Task 1.** **Visual verification** step 2 (`skills/flow/verify-and-handoff.md`)
  pipes `git -C <worktree> diff --name-only <merge-base>..HEAD` into `check-visual-trigger.sh
  --repo` for every satellite worktree, and nothing in the step says which `<merge-base>`.
  **Inline verify** in the same file already names "its merge base from this run's working notes";
  the visual step does not. KAN-659's half — the guard matching another worktree's diff — landed in
  `56260f9f`. Blast radius: one file.
- **KAN-714 — already fixed, no task.** Since `e75f19d0` (2026-09-24) every `spectre link` is
  followed by a link commit of "each `<project>/spectre/changes/<name>/link.md` the link wrote —
  the canonical worktree's and the satellite's" (`skills/flow-contracts/git-boundaries.md`,
  **Planning commits**); a satellite's `link.md` is tracked like every other.
- **KAN-715 — already fixed, no task.** Since `5b64e1a9` (2026-09-30) the panel's base sync is
  `sync-panel-base.sh`, which rebases a `MOVED`-with-no-overlap worktree itself, with no prompt
  (`skills/flow/review-panel.md`) — the wrapper entry point the ticket asked for; the classifier
  sees the script, never a bare `git rebase`.
- **KAN-718 — already fixed, no task.** `aca9b26a` (2026-09-24) and `c05173e9` (2026-09-28)
  narrowed `hooks/protect-main-checkout.py` to writes into a main checkout on its default branch,
  skipping unexpanded paths; read-only commands naming the main checkout or a `$VAR` path pass.
  The zsh word-splitting trap and the absolute/`-C` path rule are **The parent's shell
  discipline** (`skills/flow/implement.md`).
  <!-- measured: echo '{"tool_name":"Bash","tool_input":{"command":"<c>"},"cwd":"/Users/tweety53/Projects/agents"}' | python3 hooks/protect-main-checkout.py for c in 'ls $W/scripts', 'flow record pass -C /Users/tweety53/Projects/agents x', 'cat /Users/tweety53/Projects/agents/README.md | head', 'git -C $W status' @ cfb9b564 → exit 0, no deny, each -->
- **KAN-719 — already fixed, no task.** gymie-frontend `55634cf3` (2026-10-18) declares a
  `## worktree setup` line copying the main checkout's `local.properties` into a fresh worktree;
  `kickoff-worktree.sh` runs it at step 3.
  <!-- measured: git -C gymie-frontend show origin/develop:.flow/project.md | grep -A4 '## worktree setup' @ gymie-frontend 55634cf3 → the cp line -->

## Decisions

### The satellite's merge base comes from the working notes

**ID:** visual-trigger-working-notes-base
**Status:** active
**Chosen:** one sentence under **Visual verification** step 2 (`skills/flow/verify-and-handoff.md`): `<merge-base>` in both trigger calls is that worktree's own merge base from this run's working notes, never `origin/<base>` or a hand-picked ref.
**Considered:** the state file's `worktrees` map — `skills/flow/review-panel.md` already rules it out as a source, a creating run may not have written it yet; a guard refusing a match set carrying another change's `kan-*` paths — a project-specific heuristic for a defect the right base removes by construction; resolving the base inside `check-visual-trigger.sh` — the guard reads stdin paths and stays agnostic of where they came from.

### Four tickets are already fixed

**ID:** four-already-fixed
**Status:** active
**Chosen:** no task for KAN-714, KAN-715, KAN-718 or KAN-719 — each fix is in the tree, cited under **Context** with its commit.
**Considered:** re-stating each fix in a new sentence — restates what the tree already says canonically.

## Open questions
