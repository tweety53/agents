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
scripts/test-preserve-session-records.sh
scripts/test-check-unfinished-work.sh
scripts/test-check-cleanup-complete.sh
```

## lint

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
```

**There is no auto-fix command in this repository.** The Lint Fix Priority rule's "run the
auto-fix command first" step is therefore inapplicable here — not skipped. All three guards report
`file:line` and are fixed by editing the offending line, never by weakening the guard or adding a
suppression marker to silence a real hit.

**`check-finish-preflight.sh`, `preserve-session-records.sh`, `check-unfinished-work.sh` and
`check-cleanup-complete.sh` are deliberately not lint steps.** All four are `/myflow-finish` helpers
that need a change in flight and a real worktree, a repository or a state directory passed in as
arguments; they answer a question about one change, not about the state of the repository's text. A
lint step that cannot run against a bare tree would fail on every unrelated invocation, so the
omission is a decision, not an oversight. They are covered instead by their harnesses under
`## test`.

**All three guards are currently expected to exit 0.** `check-plan-provenance.sh` reports
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
