# Project configuration — authoring guidance

This file is for whoever edits a project's own `<project>/.flow/project.md`. **Project
configuration** (`skills/flow-contracts/project-configuration.md`) is canonical for what a run
resolves; nothing here is consulted by any run.

**The `.mdc` extension is what selects the shared library, and nothing else.** A project-local
`.mdc` is still nameable — write it as a path (`<project>/.cursor/rules/api.mdc`), which form 3 takes as-is.
See **The `.mdc` routing rule** (`skills/flow-contracts/project-configuration-rationale.md`)
for why.

Written out, with placeholder names standing in for a real project's own — flow's own files carry
no project's databases, ports or task names, per **Project configuration**
(`skills/flow-contracts/project-configuration.md`):

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
