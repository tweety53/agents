# Isolation: each change in its own space

## The worktree

Every change is built in its own git worktree on its own branch, so the main checkout stays clean
and usable while agents work. A hook denies edits, commits and resets on a main checkout's default
branch ([Guards and hooks](guards.md)). Cleanup removes the worktree and the branch once the change
has landed.

## The workspace

A worktree alone does not stop two changes from sharing one database. A project that declares
`## workspace isolation` gets, per change:

- **its own database**, so one change's migrations cannot reach another's schema;
- **its own storage bucket**, so two changes' uploads do not land in one place;
- **its own block of ports**, so two changes' apps can be up at the same time.

The shared services themselves are untouched: the database server, the cache and the object store
keep their containers and ports, and only the resource inside each is per change. The names derive
from the change's name alone, so there is no registry and no lock for two sessions to race on. A
new database starts empty and is filled by the project's normal migrations, never copied. Cleanup
removes the database and the bucket.

Canonical: [`skills/flow-contracts/workspace-isolation.md`](../skills/flow-contracts/workspace-isolation.md)
and, for what a project declares,
[`skills/flow-contracts/project-configuration-isolation.md`](../skills/flow-contracts/project-configuration-isolation.md).
