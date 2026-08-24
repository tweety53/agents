# State file

**A change's state record lives in PostgreSQL, owned by the `myflowd` daemon.** No command reads or
writes the database directly, and no command reads a JSON file for the live value. Every
`/myflow-*` command reaches the record through two CLI subcommands that speak HTTP to the daemon:

```bash
myflow state get [-C dir] <name>            # prints the record's JSON to stdout
myflow state set [-C dir] <name> <<<"$JSON" # reads the whole record as JSON from stdin
```

`-C dir` resolves the project key as if run from `dir` (default: the process's own working
directory) — the same purpose `git`'s own `-C` serves, and useful from a worktree or a script.

The record is keyed by **project and change name together**, so two projects may each hold a
change of the same name without collision.

**"State file" is this contract's name, not a live artifact any command still opens.** The name
survives because an on-disk JSON file — at the path below — still exists, but only as the CLI's
fallback record and the write-ahead journal's payload shape, per **The pipeline never blocks**
below. It is written only when the store could not be reached, and no command reads it while the
store answers normally.

```text
/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json
/Users/tweety53/Agents/myflow/state/<project-key>/<name>.journal
```

`<project-key>` = `<basename of main checkout>-<first 8 hex of sha1 of the main checkout's absolute path>` — e.g. `myrepo-3f9a1c02`. The basename keeps it readable; the hash makes two same-named repos in different directories unambiguous. It is the same key `myflow state get`/`set` send the daemon, so the store, the fallback file and the journal all address one record under one key.

**Resolving the main checkout is load-bearing.** `git rev-parse --show-toplevel` returns the *worktree* root when run inside a worktree, which would give apply (in a worktree) and review (in the main checkout) two different keys for the same change. Always resolve via `--git-common-dir`, which points at the **main** repo's `<project>/.git` from anywhere, including inside a worktree:

```bash
MAIN_CHECKOUT="$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd -P)"
PROJECT_KEY="$(basename "$MAIN_CHECKOUT")-$(printf '%s' "$MAIN_CHECKOUT" | shasum | cut -c1-8)"
```

This derivation is now performed inside the CLI itself (`myflow state get`/`myflow state set`
resolve it from `-C dir`, or the working directory) rather than by a skill running this recipe by
hand — but the algorithm is unchanged, and it is stated here because it is what makes the store,
the fallback file and the journal agree on one identity for one change.

**The path is resolved through symlinks — physical resolution, not the raw one — and that is
load-bearing rather than incidental.** Git records a worktree's pointer back to its main checkout
as an **already-resolved** real path, while a naive resolution of the main checkout side preserves
whatever symlinked route the operator arrived by. Without resolving both sides the same way, the
identical repository yields **two different project keys** depending on which side asks — the
exact split this section exists to prevent, reappearing one level down. It was found by running the
derivation from a real worktree whose temporary directory crossed a symlink; a test using a
constructed path would not have shown it.

Anything else deriving this key — a command, a script, or a program — resolves symlinks the same
way. A project reached through a symlinked path (a symlinked home directory, a synced folder, a
hand-made checkout symlink) otherwise splits its records between two keys the moment two mechanisms
disagree.

## The pipeline never blocks

**Every CLI path that touches the store falls back on any failure, and exits 0.** "Any failure"
means the daemon is down, the database behind it is down, the request times out, or the daemon
answers with anything other than a success — this is deliberately broader than the Jira contract's
"never a gate", because a state write happens at the end of every command: an outage that stopped
the write would strand a change at an unwritten state with the work already done.

- `myflow state get <name>`: on success, prints the store's record and exits 0. If the store
  correctly reports no record for this project and name, it prints that and **exits 1** — this is
  the store answering, not an outage, and is the one case `state get` does not fall back for. On
  every other failure it prints one line, `⚠ myflow: store unreachable — read local fallback`, then
  the on-disk fallback record if one exists — silently nothing if it does not, since there is
  nothing more honest to print — and **exits 0**.
- `myflow state set <name>`: on success, exits 0 silently. On every failure it writes the payload
  to the on-disk fallback file, appends it to the journal, prints one line,
  `⚠ myflow: store unreachable — wrote local journal`, and **exits 0**.

**A genuine refusal is the one outcome that is reported and exits non-zero**, and it is identified
by a real answer from the daemon, never by a status code alone: a response is trusted as the
daemon's own only when it carries the `Myflow-Daemon` response header, so a look-alike server
squatting the port cannot be mistaken for a genuine refusal. Under that guard, only an HTTP 409 —
the monotonic-write refusal below — is reported: `myflow state set` prints
`myflow: state set refused: ...` and **exits 1**, and the stored record is left exactly as it was.
**Every other daemon response the CLI does not treat as success — a 400 for a malformed or
undocumented-field payload included — is folded into the same "store failure" bucket as an outage**
and takes the fallback path above, exiting 0. A skill that sends a malformed payload is therefore
never stopped by this layer; the payload is written to the fallback file and journal as sent, and
what happens to it next is stated under **The journal is replayed, never merged** below.

A CLI usage error — non-JSON stdin, JSON that is not an object, stdin over the CLI's own size
cap, or a `worktrees` value that is neither `null` nor a sha (**The record** below) — is a local
input error, not a store failure: it is reported to stderr and exits 2, never falls back, and never
reaches the network.

## The record

```json
{
  "state": "IN_PROGRESS",
  "branch": "openspec/<name>",
  "worktrees": {
    "/absolute/path/to/worktree": "<merge-base sha>"
  },
  "artifactUrl": null,
  "jiraIssue": null,
  "planningEffort": null,
  "models": {
    "implementation": null,
    "reviewPanel": null,
    "panelFix": null
  },
  "reviewPanelRoster": null,
  "prUrl": null,
  "updatedAt": "2026-07-28T10:00:00Z",
  "updatedBy": "/myflow-do"
}
```

This is the whole wire shape `state get` prints and `state set` reads — the same shape the on-disk
fallback file and each journal entry's payload carry, unchanged, so replay needs no translation
step.

**`state set` also accepts, and the CLI itself injects, a `mainCheckoutPath` field** that bootstraps
the daemon's project row on a change's first write. It is transport-only: it is never part of this
record's own vocabulary, never appears in `state get`'s output, and no skill supplies it — the CLI
adds it from the same main-checkout resolution described above.

**The record's field vocabulary is closed**, and the daemon enforces it whenever it is reachable: a
payload naming a field outside those documented above (`mainCheckoutPath` aside) is rejected. Per
**The pipeline never blocks** above, that rejection is a 400, not a 409, so the CLI does not report
it as a refusal — it takes the fallback path like any other store failure. The undocumented field is
therefore not silently accepted into the store, but it is also not reported to the operator at write
time; see **The journal is replayed, never merged** for where it actually surfaces.

**Omitting a documented field from a write clears it.** `state set` sends the whole record, and the
store performs a full overwrite: a field the payload does not carry is stored as absent, exactly as
if it had been sent explicitly `null`. There is no partial-merge path — the store does not remember
what a previous write held for a field this write leaves out. This is why **Carry the record
forward on every write** below is a rule every command must follow, not a convenience: dropping a
field is how it gets erased.

- `state` — one of the three values in **States** (`skills/myflow-contracts/pipeline.md`):
  `STARTED`, `IN_PROGRESS`, `FINISHED`.
- `branch` — the change's branch, `openspec/<name>`; `null` before one exists.
- `worktrees` — an object **keyed by the absolute path** of each affected worktree, whose value is
  that worktree's merge base. `{}` when none exist or all were removed. **A `FINISHED` change may
  legitimately carry a non-empty map:** `/myflow-finish` clears only the entries whose removal
  actually succeeded, so a worktree that could not be removed stays listed and remains findable.
  See **A change spanning repositories is one record** below.

  **A value is either JSON `null` or a 40-character lowercase hexadecimal sha, and nothing else** —
  a worktree path, a short sha, an uppercase sha and an empty string are all refused. `myflow state
  set` refuses one before the store is touched: it names the offending worktree path and the
  rejected value on stderr and exits 2, writing neither the on-disk fallback file nor a journal
  entry, because a malformed merge base is a local input error in the sense of **The pipeline never
  blocks** above rather than a store outage. Journalling it instead would hide the bad value until
  the finish gate's preflight refused, which is the failure this constraint exists to stop at the
  point the value is written. The store refuses it a second time on write, with an error distinct
  from an invalid state and from a monotonic refusal, covering the one path that bypasses the CLI —
  replay of a hand-edited or out-of-band-modified fallback file; such an entry is retired from the
  journal rather than replayed forever.

  **A `null` value is legal and means *no merge base recorded* for that path** — it can occur in a
  hand-edited or out-of-band-modified fallback file. Every rule this contract and the pipeline
  already state for a **missing** recorded merge base applies to a `null` one unchanged: the
  preflight is handed `-` and refuses, the merge status is `inconclusive`, and the review command
  falls back to the staged diff. A `null` value is therefore never a licence to infer a merge base —
  it is the refusal to.
- `artifactUrl` — the published proposal artifact's URL; `null` until `/myflow-start` publishes one.
- `jiraIssue` — the key of the Jira issue driving this change (e.g. `"KAN-8"`), or `null` when no issue is linked. Written only by `/myflow-start`; every other command **carries it forward verbatim**. See **Jira integration** (`jira-integration.md`).
- `planningEffort` — the level chosen for this change's planning, or `null` when none was chosen.
  Written only by `/myflow-start`, on the run that **creates** the change; every other command
  **carries it forward verbatim**. It governs `/myflow-start`'s own reasoning depth and nothing
  else — no command derives behaviour from it, and the review panel's breadth is never scaled from
  it. The levels, and which of them is offered as the recommendation, are stated once under
  **Planning effort** (`state-file.md`) below.
- `models` — an object carrying `implementation`, `reviewPanel` and `panelFix`, each naming the
  model chosen for that role, or `null` where none was chosen. Written only by `/myflow-start`, on
  the run that **creates** the change; every other command **carries it forward verbatim**. Its
  live consumer is `/myflow-do`, which dispatches on those values. The roles, their defaults and
  how an operator override applies are stated once under
  **Model policy** (`skills/myflow-contracts/model-policy.md`), which is canonical for them; a second
  copy here is what this repository's reference guard exists to prevent. These fields record what
  was *chosen* — the SDD ledger remains the only record of what a dispatch actually ran on.

  **A record that omits `planningEffort`, `models` or `reviewPanelRoster` entirely is valid**, and
  each absent key is read as *not recorded*. Without this exception, `state get` would hand back a
  record it treats as malformed for every change written before these fields existed — a spurious
  report against a value nobody had the chance to set. The carve-out covers a key that is
  **absent**: `artifactUrl`, `jiraIssue` and `prUrl` are all *present and nullable*, which is a
  different thing from *absent*. For `reviewPanelRoster`, *not recorded* resolves to the default
  preset rather than leaving the panel unconfigured, so a command never has to ask which roster to
  use at panel time.

  **A record carrying the retired `effort` key is read as recording the equivalent level** —
  `medium` as `default`, `high` as `detailed`, `low` as `low` — and is rewritten under
  `planningEffort` on the next write that record receives. It is not treated as malformed, and the
  rewrite is **not announced as a correction**: the value was written correctly under the contract
  in force when it was set. **No migration pass is run** — no command sweeps existing records, and a
  record nothing writes to keeps the old key indefinitely without that being a fault.

  **The compatibility read is the fallback, not a preference, and every consumer performs it.** A
  command that reads only `planningEffort` reports a record recording a real level as having
  recorded none — the exact outcome this exception exists to prevent, and a promise made in prose
  and implemented nowhere is not a promise. The mechanical form is `(.planningEffort // .effort)`,
  as `/myflow-status` reads it.

  **When a record carries both keys, `planningEffort` wins.** Both are individually excepted from
  the closed-schema rule, so such a record is valid, and it needs a stated answer rather than an
  implied one. The current key is the one this contract writes and the retired one is read only for
  a record that never had it, so a record carrying both is one already migrated whose old key was
  never cleared — the new value is the later of the two. The next write drops `effort` rather than
  re-emitting it; a write renders the whole object, so leaving it out is what removes it.

  **A value outside those three reads as *not recorded*, and never makes the record malformed.**
  The retired key never held anything else under the contract in force when it was written, so a
  record carrying, say, `urgent` under it is not one this pipeline produced: it maps to no level,
  and surfacing the raw value would put a level this pipeline does not have in front of the
  operator. *Not recorded* is what remains once an unrecognisable value is discarded, and it is the
  honest answer for a value with no defined target.
- `reviewPanelRoster` — carries the review panel roster preset chosen for the change, one of
  `light`, `standard` or `full`. Written only by `/myflow-start`, on the run that **creates** the
  change; every other command **carries it forward verbatim**. Its live consumer is `/myflow-do`,
  which selects the panel's required slots and the per-task review's shape from it.
  `skills/myflow-do/SKILL.md` is canonical for what each preset means. The field is top-level rather
  than nested under `models` because a roster is not a model.
- `prUrl` — the pull request's URL once one is open; `null` otherwise. Its non-nullness is what
  records that a PR was opened, so no separate boolean exists. It is also what tells `/myflow-do`
  that a fix must be committed and pushed rather than merely staged.
- `updatedAt` — the ISO-8601 UTC instant of the last write, and **CLI-owned**: `myflow state set`
  stamps it on every write from its own clock, at full precision, overwriting whatever value the
  payload carried. The stamped instant is the one the store row, the on-disk fallback file and the
  journal entry all carry, so an entry replayed later orders by the instant its write actually
  happened at rather than the instant of the replay. No skill reads the clock for this field or
  emits it. A payload still carrying the field is accepted with its value ignored rather than
  refused. A journal entry written before this rule replays unchanged for an unrelated reason: the
  daemon decodes such an entry's own body rather than routing it back through the CLI, and that
  decoder accepts a second-precision instant and a sub-second one alike. `/myflow-status` reports
  "last update" from this field, and the store uses it to order same-state writes (see **Writes are
  monotonic in both dimensions** below).
- `updatedBy` — the command that last wrote the record, e.g. `/myflow-do`.

**This record carries no human confirmation and no fix origin.** No command observes whether the
human ran the apps, so nothing could honestly confirm that a human reviewed the work. And a fix
never moves the state, so there is no origin state for a fix to return to.

**Five fields from the twelve-stage predecessor are retired and never appear in a record this
contract produces.** A payload still carrying one is an undocumented field under the closed-schema
rule stated above:

- the field that recorded where a dynamic-target write originated — meaningless once targets
  stopped being dynamic.
- the boolean flag that recorded a change having taken the shortened single-session route — that
  route is gone, and its shape does not survive in three states. It did **not** mark a change as
  skipping review: the route it recorded ran a review panel of its own, just a smaller roster than
  the standard one. What actually skipped review was a separate *invocation flag* on the
  twelve-stage cycle, which was never a state-file field at all and so has no tombstone here.
- the cache key that pinned a stage-advance script's last-scanned commit — the script it belonged
  to no longer exists.
- the four-valued object that tracked pass/fail across four named checkpoints of the old
  twelve-stage pipeline — collapsed into the single `state` field above.
- the boolean recording whether a human had exercised the change — no command ever observes this
  honestly (see "This record carries no human confirmation" above), so the field recorded a claim
  nobody could verify.

## Writes are monotonic in both dimensions

A write is refused (HTTP 409, `myflow state set` reports and exits 1) when it would move the record
backwards **in either dimension**: to a `state` earlier in the pipeline than the one stored, or —
at the *same* `state` — to an `updatedAt` earlier than the one already recorded. The recorded
instant is the primary ordering and the pipeline state is the tiebreaker, so a replayed or
duplicated write can never silently overwrite a newer record with older field values.

The instant this ordering rests on now comes from a single writer — the CLI stamps it (`updatedAt`
under **The record** above), so every live write is ordered by one clock at one precision rather
than by two clocks whose differing precisions made a same-state write inside one second compare as
backwards.

**This same refusal covers a benign duplicate** — a write identical to one already accepted, being
retried or replayed — because the store cannot tell a superseded write from a harmless repeat of the
current one from the error alone: both carry a `state`/`updatedAt` pair no later than what is
already stored. Both are safe to retire without further action, which is exactly what
**The journal is replayed, never merged** below does with them.

## Carry the record forward on every write

Because a write renders the whole record and omission clears a field (**The record** above), every
command must first `myflow state get` the existing record and carry forward every field it does not
itself own — `artifactUrl`, `jiraIssue`, `prUrl`, `reviewPanelRoster` and `worktrees` among them —
before calling `myflow state set`. Re-emit each as read (`null` only if it was already `null`).
Dropping one erases it permanently: the published proposal link, the link to the Jira issue, the PR
(which also silently downgrades the next fix from commit-and-push to staged-only), the chosen review
panel roster preset, or the authoritative list of worktrees for a multi-repo change.

**Carrying the planning effort forward is what performs the rewrite**, and it is the one field where
*verbatim* needs saying precisely. A record read through the retired-key fallback above is carried
forward as its **mapped level under `planningEffort`** — that is the rewrite the exception promises,
and the only write that ever performs it. Re-emitting the old key instead leaves the record
permanently unmigrated, and emitting `planningEffort: null` because the new key was absent from the
read erases a level the operator chose. Every command that writes a record it did not create does
this, so no single command is responsible for a migration none of them run.

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

## A change spanning repositories is one record

A change may affect more than one repository, and the record already treats that as **one**
record: a single scalar `branch` plus the `worktrees` map above, one entry per affected repository.
The store reads the same way — a two-repository change is one record, never two.

The **project key names the project whose state directory owns the record** — it is not the list of
affected repositories. The daemon derives the affected-repository set from `worktrees` on every
write and persists it alongside the record in the same transaction, so no reader ever observes a
change with a partially updated repository set. A skill's own obligation is unchanged by any of
this: it writes `worktrees` exactly as it always has, and never supplies a repository set
separately.

The **key set of `worktrees` is the authoritative recorded list of affected worktrees** — it is
what `/myflow-finish` cleans up, and what resolves an app's root when a handoff needs an absolute
path. It is the record, not the iteration set: a step that needs "the worktrees" resolves that set
first, per **Resolving a change's worktrees** (`skills/myflow-contracts/worktree-resolution.md`), rather than
looping over this map directly. The scalar `branch` names the shared branch only. Never infer a
worktree path from a conventional layout; layout differs per repository.

```json
"worktrees": {
  "/Users/tweety53/Projects/agents-worktrees/openspec-<name>": "5ee4c9a…",
  "/Users/tweety53/Projects/other-worktrees/openspec-<name>": "b31f7c2…"
}
```

## The store starts empty

**No record predates this contract's store.** Nothing imports the JSON state files that existed
under the file-based contract, and no command reads one for a live value or as a fallback of last
resort — the on-disk file this contract still names is written only going forward, by the CLI's own
fallback path, never seeded from history. A change worked before the store existed has no record
until a command writes one for it.

## Read it, write it

```bash
myflow state get "$NAME" -C "$DIR"                    # prints the record, or falls back — see above
printf '%s' "$RECORD_JSON" | myflow state set "$NAME" -C "$DIR"
```

`state set` reads the whole object from stdin — always write every field, never a partial merge, per
**The record** and **Carry the record forward on every write** above. Never write the on-disk
fallback file or journal directly; they belong to the CLI, are machine-local, and are **never
committed, never staged, and never archived** — nothing here is part of the change.

## Planning effort

**Which file to change first.** The normative requirement — that three levels exist, that `default`
is the level offered as the recommendation, and that no level may switch a gate off — is
**Requirement: Planning effort scales the reasoning spent inside the gates, never the gates themselves** (`<agents repo>/openspec/specs/myflow-planning-effort/spec.md`).
Naming the requirement in full, rather than giving the path alone, is what makes
`<agents repo>/scripts/check-references.sh` check this pointer — an OpenSpec `### Requirement: …` heading is a
heading like any other. The guard skips a path that does not resolve, so this one is checked only
once the capability lands in `<agents repo>/openspec/specs/` at finish run 2; until then it is a reference nobody
verifies, which is said here rather than left to look otherwise. That spec is the requirement; the
table below is the
**operational form the commands read**, and it exists here so `/myflow-start` has one place to look
rather than a requirements document to interpret. Change the spec first and bring this table with
it: a table that contradicts the requirement is this file's defect, not the spec's.

Three levels, offered by `/myflow-start` on the run that creates a change, with `default` the level
offered as the recommendation:

| Level | What it changes |
|-------|-----------------|
| `low` | Questions batched rather than asked one at a time; the design presented once; `tasks.md` grouped more coarsely |
| `default` | The checklist followed with related questions grouped |
| `detailed` | Each checklist item worked separately, alternatives enumerated per open question, each design section approved on its own |

**The retired key belongs to the field, not to this table.** `effort`, its mapping onto these three
levels, which key wins when a record carries both, and what an unmapped value reads as are all
stated with the `planningEffort` field above — one statement, beside the field it governs. This
section is the levels themselves; it names where that rule lives rather than carrying a second copy
of it.

**No level may switch a gate off.** Brainstorming runs, the design approval gate holds,
writing-plans runs, and `tasks.md` is never left a thin scaffold — at every level. A lower level
means fewer rounds and coarser grouping, never a gate that does not run. A planning effort level
able to skip a gate would be a way to skip review rather than a way to size the thinking inside it.
