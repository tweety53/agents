# kan-439-check-unfinished-work-sh-false-outstanding-on — design

## Context

- `check-unfinished-work.sh` resolves a change's plan through `scripts/lib/change-plan.sh`, whose
  resolution follows a satellite `link.md` (`## Part of: <peer>:<change-id>`) to the canonical
  plan, using the caller-supplied canonical worktree first, then the `peers` file.
- On a cross-repo change whose project keeps the change directory in the canonical repo only
  (gymie's convention; its frontend repo has no spec tree), there is no `link.md` to follow and the
  canonical-worktree argument is never consulted → false `OUTSTANDING` at integrate (KAN-423,
  KAN-439).
- Run 1 already passes the canonical worktree on every call (`finish-contract-run1.md`); bf684f7
  already hard-gated the `spectre link` step with correct cwd. Full design, including the rejected
  alternatives and the testing strategy:
  `docs/superpowers/specs/2026-09-08-kan-439-check-unfinished-work-sh-false-outstanding-on-design.md`.

## Decisions

### Same-name resolution through the supplied canonical worktree

**ID:** canonical-arg-same-name-resolution
**Status:** active
**Chosen:** when no local `tasks.md` and no usable `## Part of` link exist, resolve
`<canonical-worktree>/<spec-root>/changes/<name>/tasks.md` by the same change name — no new config
key, the caller-supplied argument already names the right worktree.
**Considered:** `plan-repo:` key in `.flow/project.md` (new validated config key + containment
contract, and it still cannot name the canonical apply worktree, so worktree resolution would be
needed underneath anyway); stop at bf684f7 (leaves the hand-verification step on every treeless
cross-repo change).

### A failing `## Part of` link is never rescued by the same-name branch

**ID:** no-rescue-for-failing-part-of-link
**Status:** active
**Chosen:** the same-name branch fires only when there is no `link.md` or no `## Part of` section;
a link that exists and names its plan but does not resolve keeps failing loudly (guard exit 2).
**Considered:** rescuing any unresolvable shape (turns a structural link fault into a silent
name-match verdict).

### Signal two reads the project the plan lives in

**ID:** signal-two-reads-canonical-project
**Status:** active
**Chosen:** when the plan resolved under the canonical worktree, `flow record findings` runs with
`-C <canonical-worktree>`; the closing `flow record verdict` write stays `-C <worktree>`.
**Considered:** leaving the satellite's signal-two read as-is (a permissive `[]` under the
satellite's own project key — a CLEAR whose finding half rests on a query that cannot see the
change).

## Open questions

<!-- none — the convergence confirm closed with "Approve design, move on" -->
