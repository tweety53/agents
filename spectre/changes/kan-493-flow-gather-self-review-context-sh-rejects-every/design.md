# kan-493-flow-gather-self-review-context-sh-rejects-every

## Context

The override block in `scripts/gather-self-review-context.sh` (the `REPO_ROOT_OVERRIDE`
derivation) builds `repo_root_from_git` as the parent directory of
`git -C <root> rev-parse --git-common-dir` and demands equality with the supplied path.
`--git-common-dir` always points at the main checkout's `.git` — including when invoked from
inside a linked worktree — so a worktree argument always mismatches and exits 2 with "not the
root of a git repository". The documented caller (`skills/flow/archive.md` step 9) passes
`<landing-worktree>`: after run 2's archive step the archived change path is physically under
the landing worktree, not the main checkout, so the main checkout cannot substitute for it. The
process-cwd derivation in `validate_archived_path()` step 1 keeps `--git-common-dir` for the
opposite reason (F23: from a worktree cwd the trust anchor must still resolve to the main
checkout) and is deliberately untouched. The caller is this script's trusted invoker; accepting
a worktree root from it does not touch the untrusted-input containment
(`validate_archived_path()`), which anchors on whatever trusted root it is given.

## Decisions

### Worktree-aware validation for the `<repo-root>` override

**ID:** show-toplevel-override-validation
**Status:** active
**Chosen:** `git -C <root> rev-parse --show-toplevel` compared equal to `<root>` — one
mechanism that accepts a main checkout root and a linked worktree root and keeps refusing
everything the old check refused: a repository subdirectory (`--show-toplevel` names the
containing root), a non-repository directory (empty output), a relative path and a
symlink-divergent path (the existing lexical/real checks, unchanged and run first).
**Considered:** keeping `--git-common-dir` and special-casing worktrees (two mechanisms
negotiating one meaning of "root" — the drift the old comment block itself existed to prevent);
accepting any directory inside a repository (would silently relocate where sources 1-4 are read
from); changing the process-cwd derivation too (out of scope — F23's constraint there points the
other way).

### The code deviated from the docs, not the reverse

**ID:** code-was-wrong-not-the-doc
**Status:** active
**Chosen:** fix `scripts/gather-self-review-context.sh` and its harness only; no skill text
changes — `skills/flow/archive.md` step 9 already documents the worktree-passing shape. In-file
comments stating the old refusal (header NOTE on `<repo-root>`, `REPO_ROOT_OVERRIDE` block) are
updated in the same commit so the file does not contradict its own behavior.
**Considered:** also documenting the acceptance in skill prose (would restate what the script's
own header already carries, and no skill text says the override is main-checkout-only).

### /flow-fast state interop

**ID:** flow-fast-state-interop
**Status:** active
**Chosen:** this change runs under `/flow-fast` with the fixed decision recorded in
`.superpowers/sdd/decision.json` — execution inline, implementer skipped, a static
`primary` + `simple-reviewer` panel on one dispatch, delta rerun, compact — so a later `/flow`
resume reads a valid record instead of re-rolling.
**Considered:** none — fixed by `skills/flow-fast/SKILL.md`'s Model resolution.

## Open questions
