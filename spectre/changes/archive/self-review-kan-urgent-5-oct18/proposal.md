# self-review-kan-urgent-5-oct18

## Why

Five Jira tickets, planned as self-review findings; four hold against the trees at agents
`b61a1e6a`, spectre `c3b62de` and gymie `6a8eddc7` (`design.md`, **Context**), and half of a fifth
is already fixed:

- KAN-975: `spectre link` and `spectre validate` look for a peer's counterpart worktree only under
  `<repo>/.worktrees/<id>`, so a counterpart in flow's sibling `<repo>-worktrees/<id>` is not found
  and a cross-repo link reads one-sided.
- KAN-965: the one-sided link from a worktree is KAN-975's defect; the merge-order-only `link.md`
  refusal is already fixed.
- KAN-967: a fix-loop dispatch, whose begin the PreToolUse hook records, never gets an agent id,
  and an inline row of a run that marks no stage under its token has no window — both stay
  `Tokens: not measured`.
- KAN-973: the flake rule says to run a test until it fails, never how to make a rare flake fail
  and measure it, and no reviewer asks a flake fix for its measured proof.
- KAN-962: gymie's `verifyFeatureCatalog` checks every frontend `@Feature` code against a catalog
  that indexes only the backend's domains and `auth-ui`, so `check` is red in every gymie checkout.

## What changes

- `spectre link` and `validate` find a counterpart under `<repo>-worktrees/<id>` too.
- A fix-loop launch stamps its row by its prompt's `Key:` line, and every dispatch begin records
  the harness session it ran in, so loop and inline rows carry their tokens.
- The flake rule throttles a rare flake until it fails and measures its rate; the primary
  reviewer reports a flake fix without the measured proof.
- `verifyFeatureCatalog` checks frontend codes only in the domains the catalog holds, and the
  catalog carries every `auth-ui` code the frontend annotates.
