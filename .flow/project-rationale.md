# flow project configuration — agents — rationale

This file is the reasoning behind `.flow/project.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.** `project-get.sh`
reads only `project.md`.

## .flow/project.md — apps

This followed the
review panel passes and fix waves that found defect class after defect class in the Bash
version — canonical enumeration and full history in `check-plan-provenance.py`'s own module
docstring (this file does not restate the count, since a copied number is exactly what let an
earlier, wrong count survive six review passes). Every other guard in this repository remains Bash-only; adding Python here was a
deliberate, recorded widening of the toolchain, not a drift.

## .flow/project.md — run

`loading it is an operator step, deliberately` — because an agent left
running unattended during this change's development harvested 2,961 transcript offsets into the
database before anyone noticed.

## .flow/project.md — test

`not by a number pasted here` — which is the point:
a written duration went stale twice before this paragraph stopped carrying one.

## .flow/project.md — worktree setup

kan-389's verifier hit exactly that, and its conductor then ran
the whole `## test` list itself to get past it — a fresh-worktree fact, not a branch defect, and
the reason this key exists.

## .flow/project.md — lint

**`check-plan-shape.sh` sits beside `check-plan-provenance.sh` and `check-task-build-green.sh` in
this list but is not one of them.** Those two are project-configured — resolved through a project's
own `.flow/project.md` and run only where a project declares them — while `check-plan-shape.sh` is
**shipped**, symlinked into `skills/flow/scripts/` and cited by basename per **Guard resolution**
(`skills/flow-contracts/pipeline.md`), because the guard it protects
(`check-task-commit-fields.sh`) is itself shipped and runs in every project `/flow` touches. It
answers a bare-tree question exactly like `check-plan-provenance.sh` and `check-task-build-green.sh`
do — no arguments scans every non-archived `<spec-root>/changes/*/tasks.md` — which is why it
belongs in this list at all, for the same reason those two do.

**Its place in this list is a self-check on this repository, not how it covers the projects flow
is installed into** — and conflating the two made a permanently vacuous lint step read as
enforcement. The run here checks this repository's own `## workspace isolation` section below; a
green lint run therefore says nothing whatever about any other project's declaration. What covers
those is `/flow`, which runs this guard against
each apply worktree before it resolves the section, per **Verify** in `skills/flow/verify-and-handoff.md` —
so a declaration is validated where it is read, in whichever repository holds it. The lint entry
stays because this repository's own configuration is one more configuration worth checking, and
because it keeps the guard runnable from a bare tree.

**`check-foreign-staged.sh`, `check-finish-preflight.sh`, `check-unfinished-work.sh`,
`check-cleanup-complete.sh` and `check-worktree-processes.sh` are deliberately not lint steps.**
All five are `/flow` integrate/archive helpers that need a change in flight and a real worktree, a
main checkout, a repository or a state directory passed in as arguments; they answer a question
about one change, not about the state of the repository's text. A lint step that cannot run against
a bare tree would fail on every unrelated invocation, so
the omission is a decision, not an oversight. They are covered instead by their harnesses under
`## test`.
`check-panel-diff-size.sh`, `plan-dispatch-bundles.sh`, `check-panel-reproducers.sh`,
`check-panel-reproducer-exit-contract.sh` and
`run-reproducer.sh` are excluded for the same reason: they are `/flow` implementation helpers that
likewise need a change in flight and a worktree passed in, so they are covered by their own
harnesses under `## test` instead.

**`check-installed-citations.sh` belongs in the list for the opposite reason those are excluded.**
It takes no change-in-flight state — it derives the installed set by running a sandboxed `setup.sh`
itself, twice per invocation, rather than being handed one — so it scans the same bare tree every
other guard above does and exits identically regardless of what change, if any, is in flight. It is
the only guard in this list that shells out to the installer rather than only reading files already
on disk, which costs it real time; see the runtime note under `## test`, next to the entry its
harness added there.

**`check-dispatch-paragraphs.sh` keeps a required dispatch paragraph from silently
disappearing.** It is argument-free and self-scoped exactly like `check-guard-symlinks.sh`: the
scan root is this repository's own root, resolved from the script's own location. It checks a table
of required paragraphs — REPRODUCE, DON'T READ at `skills/flow/review-panel.md` and
`skills/flow/implement.md`, and VERBATIM REPORT — THE FACT at `skills/flow/review-panel.md` — for
each paragraph's label and a handful of load-bearing phrases per variant, held as short literals
rather than a copy of the whole blockquote, so a body reworded around those phrases passes clean. A
green run proves only that each paragraph is present at its required sites and carries its phrases —
never that any reviewer, implementer, dispatcher or fix agent actually obeyed it.

The list is cited by count nowhere in this file, deliberately: a
written count went stale the first time a guard was added to it, and the same sentence would go
stale again on the next.

`because a rule can merge and stay unreadable by every session` — which is how KAN-202's
commit-scope rule spent a day with a dangling pointer.

## .flow/project.md — self review

The report series ended at kan-380: the six changes after it all answered "No" to a prompt that
fires after `FINISHED`, when the operator has walked away, and the 30 reports before it yielded 9
Jira tickets. This key ratifies that and ends the series; set `run` to bring it back.

Recorded when the body was `skip` (3db66bf, 2026-09-05). The body has been `defer` since ff4ca7d
(2026-09-11), which saves the context bundle for `/flow-self-review` rather than ending the series.

## .flow/project.md — workspace isolation

`after `scripts/workspace.sh remove` has dropped the `database` row's resource` — the exact loss
gymie kan-468's incomplete bundle reported.
