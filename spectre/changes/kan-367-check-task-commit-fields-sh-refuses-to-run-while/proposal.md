# kan-367-check-task-commit-fields-sh-refuses-to-run-while

## Why

`check-task-commit-fields.sh` resolves which `tasks.md` a task belongs to by globbing
`<worktree>/spectre/changes/*/tasks.md` and refusing (exit 2, no verdict) whenever that glob turns
up more than one unrelated root — which is an ordinary repo state, not a pathological one, any time
two or more changes are live at once. Every `/flow` call site already knows the change name at the
point it invokes this guard; the guard just never receives it, so it falls back to guessing from the
filesystem and gives up. KAN-173's own `/flow` run hit exactly this: three live change directories
under `spectre/changes/` made the guard exit 2 on all three of that change's task commits, forcing a
hand-run fallback for every one.

## What changes

- `check-task-commit-fields.sh` accepts a new optional 6th positional argument, `<change-name>`.
  When given, it resolves that change's `tasks.md` directly — no glob, no scan of unrelated
  directories — instead of enumerating every directory under `spectre/changes/` and refusing on
  ambiguity. The existing satellite-link resolution and highest-numbered-`-fix-N`-sibling logic
  still apply, scoped to the named change's own family only.
- `implement.md`'s call site passes `<name>` (already resolved earlier in the run) as this new
  argument, eliminating the false refusal for every real `/flow` invocation.
- Omitting the argument preserves today's glob-and-refuse behavior byte for byte — the hand-run
  fallback, and every existing test, is unaffected.

Out of scope: the leftover-unarchived-directory defect this ticket also reported (`kan-297` and
`kan-363` sitting outside `spectre/changes/archive/`) turned out not to be a pipeline bug — their
archive PRs (#77, #79) were opened correctly by `/flow` run 2 and simply never merged. They, and a
third PR (#81, `kan-173`) found the same way, were merged as part of investigating this change; no
code change resulted from that half.
