## Context

- Three findings from kan-1014's self-review pass, planned per item 1 of **3. Fix every finding
  that is not `big`** (`skills/flow-self-review/SKILL.md`) on
  `self-review-kan-1014-templates-step-3-fe-make-public-statuses`, based at `7c100f40`. Each
  finding is carried by exactly one task; the brainstorm design is
  `.superpowers/sdd/2026-10-26-self-review-kan-1014-templates-step-3-fe-make-public-statuses-design.md`
  in this worktree.
- Finding 3's file lives in the gymie repository; its commit is made in
  `/Users/tweety53/Projects/gymie-worktrees/kan-1014-templates-step-3-fe-make-public-statuses` on
  `spectre/kan-1014-templates-step-3-fe-make-public-statuses` and rides that change's landing.
- Finding 2's store repair (F5 re-recorded relative) was already done by the parent.

## Decisions

### Canonical worktree as an optional argument

**ID:** verify-list-canonical-worktree

**Chosen:** an optional 8th argument `<canonical-worktree>` to `run-verify-list.sh`, used only as
the `-C` of its `flow record dispatch begin|end` calls, defaulting to `<worktree>`.

**Rejected:** deriving the canonical worktree from sibling checkouts — guesswork where the caller
already knows the path.

### Refuse an absolute reproducer at record time

**ID:** finding-reproducer-relative

**Chosen:** `validateFindingReproducer` refuses a non-`none` reproducer whose first token starts
with `/`, worded as `runreproducer.go`'s absolute-path refusal; `..` segments and later tokens stay
guard-side.

**Rejected:** teaching `check-panel-findings-closed` to resolve absolute paths — review-panel.md
already requires a relative reproducer.

### desktopTest naming is documented, not configured

**ID:** desktoptest-camelcase-doc

**Chosen:** one line in gymie's `.flow/project.md` `## lint`.

**Rejected:** excluding `desktopTest` from `FunctionNaming` — a lint-config weakening the lint rule
forbids.

### First review effort

**ID:** first-review-effort

**Chosen:** `flow-low` — three small mechanical fixes, each spelled out with its test.

### Fixer grouping

**ID:** fixer-grouping

**Chosen:** one group, Tasks 1–3, one fixer dispatch.

## Open questions

None.
