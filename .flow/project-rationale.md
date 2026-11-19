# flow project configuration — agents — rationale

This file is the reasoning behind `.flow/project.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.** `flow-guard project-get`
reads only `project.md`.

## .flow/project.md — apps

This followed the
review panel passes and fix waves that found defect class after defect class in the Bash
version — canonical enumeration and full history in the module docstring
ported into `stats/internal/guard/planprovenance.go` (this file does not restate the count, since a copied number is exactly what let an
earlier, wrong count survive six review passes). Adding Python here was a
deliberate, recorded widening of the toolchain, not a drift.

## .flow/project.md — run

`loading it is an operator step, deliberately` — because an agent left
running unattended during this change's development harvested 2,961 transcript offsets into the
database before anyone noticed.

## .flow/project.md — test

`not by a number pasted here` — which is the point:
a written duration went stale twice before this paragraph stopped carrying one.

## .flow/project.md — test: `make -s web-dist`

kan-389's verifier hit exactly that, and its conductor then ran
the whole `## test` list itself to get past it — a fresh-worktree fact, not a branch defect, and
the reason a `## worktree setup` running `make web-build` once existed. That setup ran `npm ci` and
the SPA build on every creating run, a docs-only change's included, though `go list -deps
./cmd/flow ./cmd/flow-guard` names no `internal/web` (KAN-987); the lines that need the build now
make it lazily instead.

## .flow/project.md — lint

**Its place in this list is a self-check on this repository, not how it covers the projects flow
is installed into** — and conflating the two made a permanently vacuous lint step read as
enforcement. The run here checks this repository's own `## workspace isolation` section below; a
green lint run therefore says nothing whatever about any other project's declaration. What covers
those is `/flow`, which runs this guard against
each apply worktree before it resolves the section, per **Verify** in `skills/flow/verify-and-handoff.md` —
so a declaration is validated where it is read, in whichever repository holds it. The lint entry
stays because this repository's own configuration is one more configuration worth checking, and
because it keeps the guard runnable from a bare tree.

**`flow-guard check-foreign-staged`, `flow-guard check-finish-preflight`, `flow-guard check-unfinished-work`,
`flow-guard check-cleanup-complete` and `flow-guard check-worktree-processes` are deliberately not lint steps.**
All five are `/flow` integrate/cleanup helpers that need a change in flight and a real worktree, a
main checkout, a repository or a state directory passed in as arguments; they answer a question
about one change, not about the state of the repository's text. A lint step that cannot run against
a bare tree would fail on every unrelated invocation, so
the omission is a decision, not an oversight. They are covered instead by their harnesses under
`## test`.
`flow-guard run-reproducer` is excluded for the same reason: it is a `/flow` implementation helper that
likewise needs a change in flight and a worktree passed in, so it is covered by its own
harness under `## test` instead.

The list is cited by count nowhere in this file, deliberately: a
written count went stale the first time a guard was added to it, and the same sentence would go
stale again on the next.

## .flow/project.md — workspace isolation

`after `scripts/flow-guard workspace remove` has dropped the `database` row's resource` — the exact loss
gymie kan-468's incomplete bundle reported.

## Moved by the 2026-11-13 prompt audit — incident records

Each passage below sat inline in a run-loaded file beside the rule it motivated; the rule and its
in-sentence reason stay there. The quoted fragment before each dash is the end of that rule, for
attribution; the text after it is the passage, verbatim.

### project.md — inputs: mods/flow-task-list/

- *beside the SPA's two worker threads (`stats/web/vite.config.ts`), so no suite starves another* — (kan-1052 self-review, F5: a 235 ms vitest test passed its 5 s timeout while both `-race` suites took every core)
