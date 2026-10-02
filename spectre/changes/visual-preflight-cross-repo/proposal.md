# visual-preflight-cross-repo

## Why

`check-visual-preflight` judged an app root in another git repository (a cross-repo change's sibling worktree) against `<worktree>`'s own `## workspace isolation` rows, scanned `<worktree>`'s allowed-origins lists for that app's origin, and so failed KAN-870's gymie run with 288 false `base-url`/`origins` lines. KAN-753 and KAN-868 worked around it by hand, re-running the guard from the frontend worktree with app root `.`.

## What changes

- `check-visual-preflight`: an app root inside a different git repository than `<worktree>` is checked as its own worktree — checks 2–4 against that repository's worktree root, its own `.flow/project.md` rows and `start` command, and its own apps only; check 1 (ports) and same-repo roots are unchanged.
- `skills/flow/visual-verify.md` step 3 states that a foreign root is passed as is.
