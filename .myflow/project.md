# myflow project configuration — agents

Read by globally installed myflow skills. Every key is optional; anything absent is
auto-detected from the repository instead.

## apps

This repository has **no runnable application**. It is the source of the myflow skills, commands,
and rules, installed elsewhere by `setup.sh`. There is nothing to start, no port, and no URL.

| App | Repo root | Kind | URL | Notes |
|-----|-----------|------|-----|-------|
| myflow sources | `/Users/tweety53/Projects/agents` | Bash + Python + Markdown | — | The only repo in scope. Verification is the guard scripts below plus a sandboxed `setup.sh` run. |

**This repository is Bash + Python, not Bash-only.** `scripts/check-plan-provenance.sh` is a thin
wrapper that execs `scripts/check-plan-provenance.py` (Python 3, standard library only —
`/usr/bin/python3`, no third-party imports, no pip, no network) so its fence/container classifier
can be a real block-structure parser instead of a hand-rolled Bash ERE allowlist. This followed five
review panel passes and seven fix waves that found defect class after defect class in the Bash
version — canonical enumeration and full history in `check-plan-provenance.py`'s own module
docstring (this file does not restate the count, since a copied number is exactly what let an
earlier, wrong count survive six review passes) and
`openspec/changes/kan-14-plan-provenance/design.md`'s "Post-review reshape" section. Every other
guard in this repository remains Bash-only; adding Python here was a deliberate, recorded widening
of the toolchain, not a drift.

## run

There is no service to run. To exercise the installer without touching the real home directory:

```bash
SANDBOX="$(mktemp -d)"
HOME="$SANDBOX" ./setup.sh global
```

## test

```bash
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-check-plan-provenance.sh
scripts/test-check-finish-preflight.sh
scripts/test-commit-split.sh
scripts/test-prepare-workspace.sh
scripts/test-preserve-session-records.sh
scripts/test-check-unfinished-work.sh
scripts/test-check-cleanup-complete.sh
scripts/test-gather-self-review-context.sh
scripts/test-check-task-build-green.sh
scripts/test-check-task-commit-fields.sh
scripts/test-check-workspace-isolation.sh
scripts/test-check-contract-budget.sh
scripts/test-check-vocabulary.sh
```

## lint

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/check-workspace-isolation.sh
scripts/check-contract-budget.sh
```

**There is no auto-fix command in this repository.** The Lint Fix Priority rule's "run the
auto-fix command first" step is therefore inapplicable here — not skipped. Every guard in the list
above reports `file:line` and is fixed by editing the offending line, never by weakening the guard
or adding a suppression marker to silence a real hit. The list is cited by count nowhere in this
file, deliberately: a written count went stale the first time a guard was added to it, and the same
sentence would go stale again on the next.

**`check-contract-budget.sh` is a ratchet, not a target.** It fails when a file under
`skills/myflow-contracts/`, or a `skills/*/SKILL.md` or `skills/*/SKILL-rationale.md`, outgrows the
budget declared for it in the guard's own `budgets()` table, or carries no budget at all. The table
is keyed on the path relative to the repository root, not on the bare basename, because every skill
directory has a file literally named `SKILL.md` and a basename key would collide across skills. Each
budget is the size its file had when the change that added its row landed, plus 25% — so ordinary
edits pass and a real section addition trips it, forcing a deliberate edit to the table rather than a
silent regrowth of a file every `/myflow-*` command loads. Raising a budget is the correct response
to a genuine addition; narrowing the guard's scope or deleting a row is not.

**`check-workspace-isolation.sh` is a lint step where the other `## workspace isolation` guard is
not.** It takes a project root, defaults to this repository when given none, and answers a question
about the text of `.myflow/project.md` — so it runs against a bare tree like every other guard here.
`check-cleanup-complete.sh` reads the same section and is excluded below for the opposite reason: it
needs a change in flight.

**Its place in this list is a self-check on this repository, not how it covers the projects myflow
is installed into** — and conflating the two made a permanently vacuous lint step read as
enforcement. This repository declares no `## workspace isolation` section, so the run here passes
silently and has done since the guard was added; a green lint run therefore says nothing whatever
about any project's declaration. What covers those is `/myflow-do`, which runs this guard against
each apply worktree before it resolves the section, per section 7 of `skills/myflow-do/SKILL.md` —
so a declaration is validated where it is read, in whichever repository holds it. The lint entry
stays because this repository's own configuration is one more configuration worth checking, and
because it keeps the guard runnable from a bare tree.

**`check-finish-preflight.sh`, `preserve-session-records.sh`, `check-unfinished-work.sh` and
`check-cleanup-complete.sh` are deliberately not lint steps.** All four are `/myflow-finish` helpers
that need a change in flight and a real worktree, a repository or a state directory passed in as
arguments; they answer a question about one change, not about the state of the repository's text. A
lint step that cannot run against a bare tree would fail on every unrelated invocation, so the
omission is a decision, not an oversight. They are covered instead by their harnesses under
`## test`.

**Every guard in the list is currently expected to exit 0.** `check-workspace-isolation.sh` reports
`ISOLATION-OK` and the fact that this repository declares no section; its own header carries its
full exit-code contract: 0 every project checked is well formed, 1 violations found, 2 it cannot
answer at all. `check-plan-provenance.sh` reports
"all provenance stated" — see the script's own header for the full exit-code contract: 0
clean/nothing-in-flight, 1 violations found, 2 environment, 3 containment, 4
content-classification; a caller that treats "non-zero" uniformly, as this repository's own lint
step does, is unaffected by that split. The long-standing exception recorded here previously — a
block of unattributed fenced snippets in `kan-8-myflow-updates`'s plan — cleared on its own when
that change archived, exactly as predicted, because the guard excludes `openspec/changes/archive/`
by design. There is no known exception left. A future non-zero exit is a real hit on a plan in
flight: fix the offending line by stating its provenance, never by narrowing the guard's scope or
adding a suppression marker.

## stop

This repository has **no runnable stack**, so there is nothing to stop. The key is present rather
than omitted so that `/myflow-finish`'s stack-stopped check is skipped as a recorded decision, not
as an oversight.

## standards

- `CLAUDE.md`
- `AGENTS.md`

## jira

`KAN`
