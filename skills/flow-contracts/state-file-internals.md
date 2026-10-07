# State file — internals

How the CLI and daemon behind **State file** (`skills/flow-contracts/state-file.md`) keep the record: the fallback paths, the project key, the transport field, write ordering and journal replay.

## The fallback file and the project key

**The on-disk JSON file at the path below is the CLI's fallback record and the write-ahead journal's
payload shape**, per **The pipeline never blocks** below. It is written only when the store could
not be reached, and no command reads it while the store answers normally.

```text
/Users/tweety53/Agents/flow/state/<project-key>/<name>.json
/Users/tweety53/Agents/flow/state/<project-key>/<name>.journal
```

`<project-key>` = `<basename of main checkout>-<first 8 hex of sha1 of the main checkout's absolute path>` — e.g. `myrepo-3f9a1c02`. The basename keeps it readable; the hash makes two same-named repos in different directories unambiguous. It is the same key `flow state get`/`set` send the daemon, so the store, the fallback file and the journal all address one record under one key.

**Resolving the main checkout is load-bearing.** `git rev-parse --show-toplevel` returns the *worktree* root when run inside a worktree, which would give apply (in a worktree) and review (in the main checkout) two different keys for the same change. Always resolve via `--git-common-dir`, which points at the **main** repo's `<project>/.git` from anywhere, including inside a worktree:

```bash
MAIN_CHECKOUT="$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd -P)"
PROJECT_KEY="$(basename "$MAIN_CHECKOUT")-$(printf '%s' "$MAIN_CHECKOUT" | shasum | cut -c1-8)"
```

The CLI performs this derivation itself (`flow state get`/`flow state set` resolve it from `-C
dir`, or the working directory). It is stated here because it is what makes the store, the fallback
file and the journal agree on one identity for one change.

**The path is resolved through symlinks — physical resolution, not the raw one — and that is
load-bearing rather than incidental.** Git records a worktree's pointer back to its main checkout
as an **already-resolved** real path, while a naive resolution of the main checkout side preserves
whatever symlinked route the operator arrived by. Without resolving both sides the same way, the
identical repository yields **two different project keys** depending on which side asks — the
exact split this section exists to prevent, reappearing one level down.

Anything else deriving this key — a command, a script, or a program — resolves symlinks the same
way. A project reached through a symlinked path (a symlinked home directory, a synced folder, a
hand-made checkout symlink) otherwise splits its records between two keys the moment two mechanisms
disagree.

**`state set` also accepts, and the CLI itself injects, a `mainCheckoutPath` field** that bootstraps
the daemon's project row on a change's first write. It is transport-only: it is never part of this
record's own vocabulary, never appears in `state get`'s output, and no skill supplies it — the CLI
adds it from the same main-checkout resolution described above.

## Write ordering

The instant this ordering rests on comes from a single writer — the CLI stamps it (`updatedAt`
under **The record** above) — so every live write is ordered by one clock at one precision.

**This same refusal covers a benign duplicate** — a write identical to one already accepted, being
retried or replayed — because the store cannot tell a superseded write from a harmless repeat of the
current one from the error alone: both carry a `state`/`updatedAt` pair no later than what is
already stored. Both are safe to retire without further action, which is exactly what
**The journal is replayed, never merged** below does with them.

## The journal is replayed, never merged

The on-disk journal beside the fallback file (`<name>.journal`) holds every write `state set` could
not deliver, in the order it appended them. The daemon replays it — in file order — at startup and
whenever it regains a database connection.

Each entry is a whole-record write, applied through the same monotonic rule above: conflicts resolve
by `updatedAt`, with the pipeline state as the tiebreaker, so a `FINISHED` record already in the
store is **never** overwritten by an earlier state arriving from a stale journal entry.

**An entry is retired from the journal only once its outcome is definitive**, never on any other
outcome. *Definitive* covers every case where retrying the identical entry would produce the same
result again: the write was accepted; it was refused under the monotonic rule (a genuine
supersession or a benign duplicate, per above); or it was refused for a reason that lives in the
entry's own content rather than the store's availability — an invalid `state` value, an
unresolvable project bootstrap on a change's first write, or a body that fails to decode at all,
undocumented field included. All of these leave the record correct and nothing left for that entry
to do, so all of them retire — this is where a malformed payload that the live write path let
through (as a 400 or a decode failure folded into the fallback bucket, per **The pipeline never
blocks**) actually gets resolved, rather than staying invisible.

What is *not* definitive, and so is left in the journal for the next replay, is an outcome that
might resolve differently later: a transport-shaped failure talking to the database, or the replay
being interrupted before it reaches that entry. Replay stops at the first such entry in a given
journal file, which is what makes an interrupted replay repeat rather than lose work — but it does
not stop for an entry whose refusal is already known to be permanent. If the daemon stops partway
through a replay, the entries not yet resolved simply remain, and the next replay processes them
without duplicating the ones already applied.

## A change spanning repositories

```json
"worktrees": {
  "/Users/tweety53/Projects/agents-worktrees/<name>": "5ee4c9a…",
  "/Users/tweety53/Projects/other-worktrees/<name>": "b31f7c2…"
}
```
