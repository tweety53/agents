# State file — rationale

This file is the reasoning behind `skills/flow-contracts/state-file.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.**

## state-file.md — The name

**"State file" is this contract's name, not a live artifact any command still opens.** The name survives because an on-disk JSON file — at the path below — still exists, but only as the CLI's fallback record and the write-ahead journal's payload shape, per **The pipeline never blocks** (`skills/flow-contracts/state-file.md`).

## state-file.md — Physical path resolution

`reappearing one level down` — it was found by running the derivation from a real worktree whose temporary directory crossed a symlink.

## state-file.md — `artifactUrl`

`artifactUrl` is kept in the record rather than dropped so a change created before that decision (`publish-proposal-removed`), whose `artifactUrl` is still populated, is not read as malformed.

## state-file.md — Planning effort

`planningEffort` is a legacy field. No run writes it, and a level recorded on a change created
before the question was retired governs nothing.

**No gate is ever switched off.** Brainstorming runs, the design approval gate holds,
writing-plans runs, and `tasks.md` is never left a thin scaffold.

## state-file.md — The record

**This record carries no human confirmation and no fix origin.** No command observes whether the
human ran the apps, so nothing could honestly confirm that a human reviewed the work. And a fix
never moves the state, so there is no origin state for a fix to return to.

## state-file.md — The on-disk file is written, never seeded

The on-disk fallback file is written only by the CLI's own fallback path. No command reads a JSON
file it did not write there, and nothing is imported into the store from history. A change with no
record in the store has none until a command writes one.
