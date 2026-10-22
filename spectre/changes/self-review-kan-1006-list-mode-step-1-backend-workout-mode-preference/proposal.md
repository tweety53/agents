# self-review-kan-1006-list-mode-step-1-backend-workout-mode-preference

## Why

kan-1006's self-review pass found four defects, none `big`:

- **F-A (flow-fix).** Slot prompts never carry the reproducer authoring contract:
  `stats/internal/guard/renderslotprompt.go:160` renders only TOOLS, FOREGROUND BUILDS, NO
  DELEGATION, REPRODUCE, DON'T READ and WORKTREES, and `skills/flow/review-panel.md` asks the rule
  carried on every slot's dispatch prompt without a verbatim paragraph to carry. kan-1006's
  rendered slot prompt said only "a runnable command … or `none — <reason>`"; 4 of 5 reproducers
  came back with prose `# demonstrates:` / `# premise:` lines the parent rewrote, and one
  `n=$(… | grep -c zzz)` under `set -euo pipefail` exited 1 on zero matches, reading a zero count as
  defect demonstrated.
- **F-B (flow-fix, gymie dev tooling).** gymie's before-figure recipe (`.flow/project.md`) says
  "then stop the before backend" without how; `./gradlew --stop`, the reflex, stops every daemon of
  that Gradle version, the change's own stack among them (this run; kan-754's self-review before it).
- **F-C (flow-fix).** A `git checkout --` / `git restore <path>` on an uncommitted mutated file during
  a mutation proof nearly destroyed an edit, despite MUTATION PROOF in `skills/flow/review-panel.md`;
  nothing enforces it.
- **F-D (flow-stats-app).** kan-1006's dispatch rows 6966 (`panel-1-primary+principles`) and 6967
  (`panel-1-test-sanity+call-efficiency`) carry no agent id and empty metrics — the
  stamp-before-row race 90efd086 fixed forward.

## What changes

- `skills/flow/review-panel.md` carries a verbatim `> **REPRODUCER:**` dispatch paragraph;
  `render-slot-prompt` renders it into every slot prompt and `check-dispatch-paragraphs` requires
  it with its load-bearing phrases.
- gymie's before-figure recipe stops the before backend by its port listener, never
  `./gradlew --stop`.
- `hooks/protect-main-checkout.py` denies, in a linked worktree, a `git checkout … -- <path>` or
  `git restore <path>` naming a path with unstaged edits, pointing at `break-and-prove.sh` and
  `mutate-and-verify.sh`; `skills/flow-contracts/git-boundaries.md` states it under
  **Restoring a path with unstaged edits**.
- Rows 6966 and 6967 are restamped from their transcripts with `flowd restamp-dispatch`.
