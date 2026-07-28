# myflow project configuration — agents

Read by globally installed myflow skills. Every key is optional; anything absent is
auto-detected from the repository instead.

## apps

This repository has **no runnable application**. It is the source of the myflow skills, commands,
and rules, installed elsewhere by `setup.sh`. There is nothing to start, no port, and no URL.

| App | Repo root | Kind | URL | Notes |
|-----|-----------|------|-----|-------|
| myflow sources | `/Users/tweety53/Projects/agents` | Bash + Markdown | — | The only repo in scope. Verification is the guard scripts below plus a sandboxed `setup.sh` run. |

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
scripts/test-state-advance.sh
```

## lint

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
```

**There is no auto-fix command in this repository.** The Lint Fix Priority rule's "run the
auto-fix command first" step is therefore inapplicable here — not skipped. Both guards report
`file:line` and are fixed by editing the offending line, never by weakening the guard or adding a
suppression marker to silence a real hit.

## standards

- `CLAUDE.md`
- `AGENTS.md`

## jira

`KAN`
