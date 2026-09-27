# kan-798-flow-fix-protect-main-checkout-py-false

## Why

`hooks/protect-main-checkout.py` denied three legitimate commands during kan-741's integrate run
(filed as KAN-798 from that run's deferred self-review): a redirect target with an unexpanded
`$C/...` variable, a landing-worktree path that did not exist yet at check time, and a `cp` whose
argument scan ran across newlines into later lines. False positives from a hygiene guard train
exactly the workaround habit the guard exists to prevent — rewriting commands to appease a scanner.

Reachability at base `4a278320`, reproduced 2026-09-28: the mode-2 denial (cp into
`<main>/.worktrees/<new>/…`) and the mode-3 denial (multi-line `cp` swallowing a later line's
`git add <main>/README.md` argument) still reproduce; mode 1's denial is already gone —
`resolve()` lets `$`/backtick paths through — but the let-through leaves a false negative
(`C=<main>; echo x > $C/f.txt` is not caught), which the issue's prescribed expansion fixes.

## What changes

- Command text is pre-split into logical lines before tokenizing and each line is scanned
  separately: no writer scan crosses a line boundary; `cd` state threads across lines as bash does.
- Pure `NAME=value` assignments in the command text are collected and `$NAME`/`${NAME}` expanded in
  tokens before path resolution; tokens still carrying `$` or a backtick stay let-through.
- A path whose nearest existing ancestor is a `.worktrees` directory is future worktree content,
  outside the protected tree; a new file anywhere else in the main checkout still denies.
- `scripts/test-protect-main-checkout.sh` gains nine cases pinning the three behaviors and their
  review-hardened edges (26–34, plus the fix round's 35–37).
