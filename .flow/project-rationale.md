# flow project configuration — agents — rationale

This file is the reasoning behind `.flow/project.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.** `flow-guard project-get`
reads only `project.md`.

## .flow/project.md — apps

Canonical enumeration and full history in the module docstring
ported into `stats/internal/guard/planprovenance.go` (this file does not restate the count). Adding Python here was a
deliberate, recorded widening of the toolchain, not a drift.

## .flow/project.md — run

`loading it is an operator step, deliberately` — because an agent left running unattended can harvest transcript offsets into the database.

## .flow/project.md — test

`not by a number pasted here` — which is the point: a written duration goes stale.

## .flow/project.md — test: `make -s web-dist`

A `## worktree setup` running `make web-build` ran `npm ci` and
the SPA build on every creating run, a docs-only change's included, though `go list -deps ./cmd/flow ./cmd/flow-guard` names no `internal/web`; the lines that need the build now
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

**`flow-guard check-foreign-staged`, `flow-guard check-finish-preflight` and `flow-guard check-unfinished-work`
are deliberately not lint steps.**
All three are `/flow` integrate/cleanup helpers that need a change in flight and a real worktree, a
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
