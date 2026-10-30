# self-review-kan-1060-flow-cost-visual-verify-now-costs-more-than

## Why

The self-review pass of `/flow` change `kan-1060-flow-cost-visual-verify-now-costs-more-than` found
three non-`big` pipeline defects:

1. `pipeline-fix.sh verify` runs no visual suite when a loop's diff touches a `ui paths` glob, so a
   pipeline-fix or self-review loop lands a UI change on the default branch with stale baselines:
   `084f060f` landed `stats/web/src/views/RunDetail.tsx` that way, and the next run's integrate
   stopped on the visual gate and re-recorded four baselines (`dbb6b4f1` was the same class).
2. On the gated reviewer's on-top fix route, `skills/flow/gated-review-fix.md` says the guard's
   re-run "covers nothing", so the fix commit's paths are never field-checked: the run's task-12
   on-top fix commit added a test file outside task 12's `**Files:**`, caught only at inline verify
   after two blocked verify rows (about 21 min).
3. Under zsh, `--app $W:<root>=<url>` is mangled by the `:s` history modifier, and
   `visual-verify-pass.sh prepare` exits 2 with a usage line naming neither the cause nor the cure.

## What changes

- `pipeline-fix.sh verify` also runs the project's `## visual verification` `verify` row, resolved
  as `run-verify-list.sh` resolves it, when the branch's diff touches a `ui paths` glob, printing
  its `green:` or `red:` line; `land`, which always verifies, then refuses a red visual suite.
- `skills/flow/gated-review-fix.md` has the parent run `check-task-commit-fields.sh` on the on-top
  fix commit before the re-review is dispatched.
- `visual-verify-pass.sh prepare`'s `--app` usage error names the zsh cause and the `${W}` cure.
