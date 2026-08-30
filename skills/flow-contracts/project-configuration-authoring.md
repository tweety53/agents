# Project configuration — authoring guidance

This file is for whoever edits a project's own `<project>/.flow/project.md`. **Project
configuration** (`skills/flow-contracts/project-configuration.md`) is canonical for what a run
resolves; nothing here is consulted by any run.

**The `.mdc` extension is what selects the shared library, and nothing else.** A project-local
`.mdc` is still nameable — write it as a path (`<project>/.cursor/rules/api.mdc`), which form 3 takes as-is.
See **The `.mdc` routing rule** (`skills/flow-contracts/project-configuration-rationale.md`)
for why.

Written out, with placeholder names standing in for a real project's own — this file carries no
project's databases, ports or task names, per its opening paragraph:

```markdown
## workspace isolation

| Resource | Variable | Default | In a workspace |
|----------|----------|---------|----------------|
| `database` | `DB_URL` | `jdbc:postgresql://localhost:5432/appdb` | `jdbc:postgresql://localhost:5432/appdb_<id_underscored>` |
| `bucket` | `MEDIA_BUCKET` | `appdb-media` | `appdb-media-<id>` |
| `cache index` | `CACHE_INDEX` | `0` | `probed` |
| `port` | `API_PORT` | `8080` | `+<offset>` |
| `port` | `WEB_PORT` | `3000` | `+<offset>` |
| `url` | `MEDIA_BASE_URL` | `http://localhost:9000/appdb-media` | `http://localhost:9000/<value:MEDIA_BUCKET>` |
| `url` | `WEB_URL` | `http://localhost:3000` | `http://localhost:<value:WEB_PORT>` |
| `url` | `DB_CONSOLE_URL` | `http://localhost:8081/?db=appdb` | `http://localhost:8081/?db=appdb_<id_underscored>` |

| Command | Runs |
|---------|------|
| `create` | `./scripts/workspace create` |
| `remove` | `./scripts/workspace remove` |
| `survivors` | `./scripts/workspace survivors` |
```

See **The three-`url`-rows worked example**
(`skills/flow-contracts/project-configuration-rationale.md`) for a walkthrough of the three
`url` rows in the worked example above.

**Which of these rules a script checks, and which are left to the agent.** The distinction matters
because the two failure modes are not alike: a rule a script checks fails the same way on every run,
and a rule an agent applies is re-performed from this text each time and can be skipped without
anything erroring. So the split is written down rather than left to be inferred from a passing lint
run — a reader who assumed the whole of this section were checked would trust a green run further
than it goes.
