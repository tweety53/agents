# Agent Instructions (Antigravity)

This file is the active instruction set for Antigravity sessions in the `agents` repository.
Project-specific operational rules match `CLAUDE.md` and `.flow/project.md`.

---

## Mandatory Rules

### Lint Fix Priority

Run auto-format first, then every lint check declared in `.flow/project.md`'s `## lint` section before claiming work done:

```bash
cd stats && gofmt -w .
scripts/flow-guard check-guard-calls
cd stats && make -s web-dist && go vet ./... && gofmt -l .
cd stats && make -s web-dist && cd web && npx tsc -b
```

### No incident history in instruction files

Rules, skills, contracts, templates and rationale appendices state the rule and its reason, never the incident behind it; incidents are recorded with `flow record incident`.

### No scripts in bash

Every new script, and any new logic added to an existing one, is Go in `flow-guard`:
`stats/internal/guard/<name>.go`, registered in `guard.Registry`, tested beside it, called as
`flow-guard <name>`. `setup.sh` and `scripts/flow-guard` are the only bash files in this
repository and stay that way.

### Never stop the dev workspace's stats service or its storage

`flowd` on `127.0.0.1:4173`, the `flow-postgres` container on host port 5433, and the default
`flow` database inside it are the dev workspace's shared store. Never stop or drop them
(`docker compose down`, `launchctl unload`, `kill`). Per-worktree derived resources are managed
by `scripts/flow-guard workspace create|remove <id>`.

### Reinstall after a change to this repository lands

From the main checkout:

```bash
./setup.sh global
./setup.sh antigravity
make -C stats restart
```
