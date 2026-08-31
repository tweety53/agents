# kan-373-archive-step-follows-landing-route-default

## Why

`skills/flow-contracts/finish-contract-run2.md` (step 10) and `skills/flow/archive.md` state
unconditionally that the archive commit (`chore(spectre): archive <name>`) always lands via a
pull request, even when the change's own branch was landed with "merge and push" — a direct local
merge and push chosen at run 1's landing question. Every `/flow` run that picks merge-and-push
therefore still ends with one more small PR to merge by hand; observed three runs in a row on this
repository. Separately, every run re-asks the landing question with "Open a pull request" as the
default recommendation, with no way for a repository to record that it always wants a different
answer.

## What changes

**Archive follows the parent route, same-invocation only.** When `skills/flow/archive.md` runs as
the direct continuation of `skills/flow/integrate.md`'s merge-and-push route (same invocation,
same session token — see `skills/flow/integrate.md`'s "After merge-and-push specifically"), its
step 10 pushes `chore/archive-<name>`, merges it into `<base>`, and pushes `<base>` — the same
three sub-steps run 1's own merge-and-push route already performs, applied to the archive branch
instead. A standalone bare `/flow <name>` invocation that reaches run 2 later — the only way the
PR or manual routes ever reach archiving — has no memory of the original route and keeps opening a
PR for the archive commit exactly as today; that is already the correct, conservative behavior for
a route this pipeline cannot reconstruct across invocations.

**A new optional project-configuration key, `## default landing route`.** Declared in
`<project>/.flow/project.md` as a single-line body, one of `pull request` / `merge and push` /
`manual`. `skills/flow/integrate.md`'s landing question uses it as that prompt's own
`(default, recommended)` option in place of always defaulting to "Open a pull request" — the
question is still always asked, per the pipeline's no-flags, no-silent-skip guardrails; only the
recommended option changes. Absent, or a body that fails to match one of the three literals
exactly, is reported by name and dropped, falling back to today's behavior — the same pattern
`## jira` and `## standards` already use for a malformed row.

This repository's own `<project>/.flow/project.md` sets `## default landing route` to
`merge and push` as part of this same change, since this repo has used that route the last two
runs in a row.
