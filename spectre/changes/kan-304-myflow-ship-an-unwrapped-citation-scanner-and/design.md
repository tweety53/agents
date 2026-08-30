## Context

`scripts/check-references.sh` (KAN-301) already scans this repo for a citation that names a moved
section but still points at the old file. It runs project-wide under `## lint`, but only at
`/flow`'s `flow.verify` stage — after the review panel has already dispatched, so reviewers never
see it. KAN-304, filed by self-review of KAN-295 (angle: what could be automated), asks to surface
this before the panel rather than after it.

## How

### New project-config key: `## review panel citation check`

Optional, single fenced command block — same shape as `## stop`:

```markdown unverified:this exact shape confirmed against the ## stop key when task 5 writes it
## review panel citation check

\`\`\`bash
scripts/check-references.sh
\`\`\`
```

Documented in `skills/flow-contracts/project-configuration.md` alongside the other keys. Absent
key → the whole step is skipped, exactly like every other optional key.

### Trigger: `check-panel-citation-trigger.sh <worktree> <merge-base>`

Decides *whether* to run the configured command — a diff-touched-Markdown test, project-agnostic,
no project-config read of its own:

- Exit 0 — the diff (`git diff --name-only <merge-base>`, staged + unstaged) touches at least one
  `.md` or `.mdc` path.
- Exit 1 — it does not.
- Exit 2 — usage error: missing argument, `<worktree>` not a directory, or the diff cannot be read.

Mirrors `check-visual-trigger.sh`'s split of "decide whether" (a script) from "what to do" (prose)
— the established pattern for every mechanical pipeline decision in this repo, so this one is not a
new shape to learn.

### Wiring: `skills/flow/review-panel.md`

Inside the existing `flow.review-panel` stage — no new stage key, since this is a preparatory step
like the dispatch-context-bundle rebuild already there:

1. Run `check-panel-citation-trigger.sh <worktree> <merge-base>`.
2. On exit 0, read `.flow/project.md`'s `## review panel citation check` key. If declared, run its
   command from the apply worktree, capture combined stdout+stderr verbatim to
   `<worktree>/.superpowers/sdd/citation-check.md`.
3. On exit 1, or the key absent, skip silently — no file written, no paragraph added.
4. Every panel slot's dispatch prompt gains a new **CITATION CHECK** paragraph, present only when
   step 2 wrote a file, pointing at that file's path alongside `final-review.diff`.

**Never blocks.** The configured command's exit code is irrelevant to whether panel dispatch
proceeds — a non-zero exit (stale citations found) is still just captured output for reviewers to
act on if they choose. The actual gate remains `flow.verify`'s existing `## lint` run, unchanged.

### This repo's own configuration

`.flow/project.md` declares:

```markdown unverified:this exact shape confirmed against the ## stop key when task 5 writes it
## review panel citation check

\`\`\`bash
scripts/check-references.sh
\`\`\`
```

No new scanning algorithm — `scripts/check-references.sh` is reused as-is.

## Decisions

### Reuse check-references.sh rather than shipping a second script

**ID:** reuse-check-references
**Status:** active
**Chosen:** Wire the existing guard into panel dispatch — no new scanning algorithm.
**Considered:** Ship `scripts/citation-scan.sh` as the ticket literally names it — rejected as
duplicating check-references.sh's unwrap + section/path matching + historical-dir exclusion logic,
against this repo's own "no duplication" stance.

### Gate on the diff touching `.md`/`.mdc`, not on detecting an actual section move

**ID:** gate-on-markdown-touch
**Status:** active
**Chosen:** A cheap extension check (`check-panel-citation-trigger.sh`) — the ~34s whole-repo scan
only runs when the diff could plausibly matter.
**Considered:** Always run it unconditionally — rejected as wasted cost on every pure-code
implementation run. Detect an actual section relocation (heading removed from one file, added
verbatim to another) — rejected as by far the most complex option for a marginal gain over the
extension check, and no simpler even after reuse.

### A new project-config key, not a hardcoded script name in review-panel.md

**ID:** project-config-key-not-hardcode
**Status:** active
**Chosen:** `## review panel citation check` in `.flow/project.md`, resolved the same way `## stop`
and `## lint` are.
**Considered:** Hardcode `check-references.sh` by name inside `review-panel.md` — rejected outright:
`review-panel.md` is a globally-installed skill that runs in any project, and project-configuration.md's
opening rule is explicit that flow "must never carry one project's apps, ports, task names, or
credentials in its own files."

### No new `flow.*` stage key

**ID:** no-new-stage-key
**Status:** active
**Chosen:** The citation check is a preparatory step inside the existing `flow.review-panel` stage.
**Considered:** A new `flow.panel-precheck` stage — rejected as unnecessary ceremony for a step with
no human gate of its own, exactly the same reasoning that already keeps the dispatch-context-bundle
rebuild un-staged within `flow.review-panel`.

## Open questions

None — the design converged in one round with no deferred question.
