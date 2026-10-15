# self-review-kan-947-frontend-sends-duplicate-and-wasted-reads-on

## Why

The self-review pass over kan-947 found five defects, none `big`:

- `run-reproducer.sh`'s 20-second bound kills a reproducer that builds or runs tests (a Gradle
  suite). `RUN_REPRODUCER_BOUND_SECONDS` already raises it, but the comment above its read says it
  "can only ever shrink the wait" — false, the value is unclamped — and no run-loaded document
  says to set it.
- In a cross-repo run the WORKTREES paragraph prefixes a finding's `file:line` with its repository
  name but says nothing of a reproducer's `# demonstrates:`/`# premise:` paths, so a reproducer
  citing another worktree's file cannot resolve, although the exit-contract guard already resolves
  the `<repo>:` prefix.
- `/flow`'s **Reading the state** reads exit 1 as no state; a `flow state get` chained after
  another command that failed was misread as a creating run.
- The visual trigger never fires for a frontend-only cross-repo change: each worktree's diff is
  matched only against its own repository's `ui paths`, and gymie's names the frontend's files as
  `gymie-frontend/**` in the canonical project.
- Inline verify runs a worktree's whole `## lint` and `## test` even when that worktree's diff is
  only its spectre change directory.

## What changes

- `runreproducer.go`'s comment states the override is unclamped; `review-panel-fix-round.md` tells
  every run of a test-building reproducer to carry `RUN_REPRODUCER_BOUND_SECONDS`.
- The WORKTREES paragraph prefixes `# demonstrates:` and `# premise:` paths citing a non-canonical
  worktree with that worktree's repository name.
- **Reading the state** runs `flow state get` as its own Bash call and names the no-state line.
- `check-visual-trigger.sh` takes `--repo <name>`, matching each path as `<name>/<path>`; in a
  cross-repo run every other worktree's diff is also matched that way against the canonical
  worktree's `ui paths`, in `flow.visual-verify` and in the parent's full-suite visual run.
- Inline verify skips, in a multi-worktree run, a worktree with nothing outside its change directory.
