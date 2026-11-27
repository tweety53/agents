# Integrate — a change to the agents repository's own tooling

Loaded by `skills/flow/integrate.md`'s load directive, only when the canonical repository is
`<agents repo>` and the branch's diff names a path under `scripts/`, `stats/` or `skills/`.

**A change to `<agents repo>`'s own tooling integrates with the branch's copy of it.** When the
canonical repository is `<agents repo>` and `git -C <canonical-worktree> diff --name-only
<base>...HEAD` names a path under `scripts/`, `stats/` or `skills/` — the three trees whose
installed copies a run executes — every `flow-guard integrate-change` call this file and
`skills/flow/cleanup.md` make resolves to the branch's own copy and runs with a `flow` built from
the branch, both first on `PATH`:

```bash
cd <canonical-worktree>/stats && go build -o <scratch>/bin/flow ./cmd/flow
PATH=<scratch>/bin:<canonical-worktree>/scripts:$PATH flow-guard integrate-change prepare <main-checkout> <name> --session-token mf-<literal-token>
```

The installed script and CLI predate the change, and the script builds its guards from its own
checkout (`<agents repo>/scripts/flow-guard`) while taking `flow` from `PATH`.
