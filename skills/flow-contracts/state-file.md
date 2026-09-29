# State file

**A change's state record lives in PostgreSQL, owned by the `flowd` daemon.** No command reads or
writes the database directly, and no command reads a JSON file for the live value. Every
`/flow*` command reaches the record through two CLI subcommands that speak HTTP to the daemon:

```bash
flow state get [-C dir] <name>            # prints the record's JSON to stdout
flow state set [-C dir] <name> <<<"$JSON" # reads the whole record as JSON from stdin
```

`-C dir` resolves the project key as if run from `dir` (default: the process's own working
directory) — the same purpose `git`'s own `-C` serves, and useful from a worktree or a script.

The record is keyed by **project and change name together**, so two projects may each hold a
change of the same name without collision.

## The pipeline never blocks

**Every CLI path that touches the store falls back on any failure, and exits 0.** "Any failure"
means the daemon is down, the database behind it is down, the request times out, or the daemon
answers with anything other than a success.

- `flow state get <name>`: on success, prints the store's record and exits 0. If the store
  correctly reports no record for this project and name, it prints that and **exits 1** — this is
  the store answering, not an outage, and is the one case `state get` does not fall back for. On
  every other failure it prints one line, `⚠ flow: store unreachable — read local fallback`, then
  the on-disk fallback record if one exists — silently nothing if it does not, since there is
  nothing more honest to print — and **exits 0**.
- `flow state set <name>`: on success, exits 0 silently. On every failure it writes the payload
  to the on-disk fallback file, appends it to the journal, prints one line,
  `⚠ flow: store unreachable — wrote local journal`, and **exits 0**.

**A genuine refusal is the one outcome that is reported and exits non-zero**, and it is identified
by a real answer from the daemon, never by a status code alone: a response is trusted as the
daemon's own only when it carries the `Flow-Daemon` response header, so a look-alike server
squatting the port cannot be mistaken for a genuine refusal. Under that guard, only an HTTP 409 —
the monotonic-write refusal below — is reported: `flow state set` prints
`flow: state set refused: ...` and **exits 1**, and the stored record is left exactly as it was.
**Every other daemon response the CLI does not treat as success — a 400 for a malformed or
undocumented-field payload included — is folded into the same "store failure" bucket as an outage**
and takes the fallback path above, exiting 0. A skill that sends a malformed payload is therefore
never stopped by this layer; the payload is written to the fallback file and journal as sent, and
what happens to it next is stated under **The journal is replayed, never merged**
(`skills/flow-contracts/state-file-internals.md`).

A CLI usage error — non-JSON stdin, JSON that is not an object, stdin over the CLI's own size
cap, or a `worktrees` value that is neither `null` nor a sha (**The record** below) — is a local
input error, not a store failure: it is reported to stderr and exits 2, never falls back, and never
reaches the network.

## The record

```json
{
  "state": "IN_PROGRESS",
  "branch": "spectre/<name>",
  "worktrees": {
    "/absolute/path/to/worktree": "<merge-base sha>"
  },
  "artifactUrl": null,
  "jiraIssue": null,
  "planningEffort": null,
  "models": {
    "default": null
  },
  "prUrl": null,
  "updatedAt": "2026-07-28T10:00:00Z",
  "updatedBy": "/flow"
}
```

This is the whole wire shape `state get` prints and `state set` reads — the same shape the on-disk
fallback file and each journal entry's payload carry, unchanged, so replay needs no translation
step.

**The record's field vocabulary is closed**, and the daemon enforces it whenever it is reachable: a
payload naming a field outside those documented above (`mainCheckoutPath` aside) is rejected. Per
**The pipeline never blocks** above, that rejection is a 400, not a 409, so the CLI does not report
it as a refusal — it takes the fallback path like any other store failure. The undocumented field is
therefore not silently accepted into the store, but it is also not reported to the operator at write
time; see **The journal is replayed, never merged**
(`skills/flow-contracts/state-file-internals.md`) for where it actually surfaces.

**Omitting a documented field from a write clears it.** `state set` sends the whole record, and the
store performs a full overwrite: a field the payload does not carry is stored as absent, exactly as
if it had been sent explicitly `null`. There is no partial-merge path — the store does not remember
what a previous write held for a field this write leaves out. This is why **Carry the record
forward on every write** below is a rule every command must follow, not a convenience: dropping a
field is how it gets erased.

- `state` — one of the three values in **States** (`skills/flow-contracts/pipeline.md`):
  `STARTED`, `IN_PROGRESS`, `FINISHED`.
- `branch` — the change's branch, `spectre/<name>`; `null` before one exists.
- `worktrees` — an object **keyed by the absolute path** of each affected worktree, whose value is
  that worktree's merge base. `{}` when none exist or all were removed. **A `FINISHED` change may
  legitimately carry a non-empty map** — per **Run 2 — the branch is merged**
  (`finish-contract-run2.md`) step 8, a worktree that could not be removed stays listed and findable.
  See **A change spanning repositories is one record** below.

  **A value is either JSON `null` or a 40-character lowercase hexadecimal sha, and nothing else** —
  a worktree path, a short sha, an uppercase sha and an empty string are all refused. `flow state
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
- `artifactUrl` — `null` for the life of the change: `/flow` publishes no proposal artifact
  (design.md's `publish-proposal-removed`), so nothing ever writes this field a non-null value.
- `jiraIssue` — the key of the Jira issue driving this change (e.g. `"KAN-8"`), or `null` when no issue is linked. Written only on the run that **creates** the change; every other invocation **carries it forward verbatim**. See **Jira integration** (`jira-integration.md`).
- `planningEffort` — a legacy field: the level recorded for an older change's planning, or `null`.
  No run writes it; every invocation **carries it forward verbatim**. It governs nothing — no
  command derives behaviour from it, and the review panel's breadth is never scaled from it.
- `models` — a legacy field: an object carrying one field, `default`, naming the model recorded
  for an older change, or `null`. The run that **creates** the change writes it `null` — no run
  asks a model question — and every other invocation **carries it forward verbatim**. It governs
  nothing: `/flow` dispatches on the model it resolves per run. The model's default and how an operator override applies are
  stated once under **Model resolution** (`skills/flow/SKILL.md`), which is canonical
  for them; a second copy here is what this repository's reference guard exists to prevent. This
  field records what was *chosen* — the SDD ledger remains the only record of what a dispatch
  actually ran on.

  **A record that omits `planningEffort` or `models` entirely is valid**, and each absent key is
  read as *not recorded*. The carve-out covers a key that is **absent**: `artifactUrl`,
  `jiraIssue` and `prUrl` are all *present and nullable*, which is a different thing from *absent*.

- `prUrl` — the pull request's URL once one is open; `null` otherwise. Its non-nullness is what
  records that a PR was opened, so no separate boolean exists. It is also what tells `/flow`
  that a fix must be committed and pushed rather than merely staged.
- `withdrawn` — the withdrawal marker: `true` on a record the withdrawal route terminated — a
  change abandoned before planning, closed `FINISHED` without an archive (**The withdrawal
  route**, `skills/flow/withdrawal.md`). Absent or `false` on every record any other route wrote,
  and refused by the store paired with any state other than `FINISHED`
  (`store.ErrInvalidState`). The CLI carries the field byte-transparently — `state set` validates
  the object and the worktree values and forwards the body — so no CLI change writes it; `state
  get` prints it as stored.
- `updatedAt` — the ISO-8601 UTC instant of the last write, and **CLI-owned**: `flow state set`
  stamps it on every write from its own clock, at full precision, overwriting whatever value the
  payload carried. The stamped instant is the one the store row, the on-disk fallback file and the
  journal entry all carry, so an entry replayed later orders by the instant its write actually
  happened at rather than the instant of the replay. No skill reads the clock for this field or
  emits it. A payload still carrying the field is accepted with its value ignored rather than
  refused. A journal entry replays through the daemon's own decoder
  rather than back through the CLI, and that decoder accepts a second-precision instant and a
  sub-second one alike. `/flow-status` reports
  "last update" from this field, and the store uses it to order same-state writes (see **Writes are
  monotonic in both dimensions** below).
- `updatedBy` — the command that last wrote the record, always `/flow`.

## Writes are monotonic in both dimensions

A write is refused (HTTP 409, `flow state set` reports and exits 1) when it would move the record
backwards **in either dimension**: to a `state` earlier in the pipeline than the one stored, or —
at the *same* `state` — to an `updatedAt` earlier than the one already recorded. The recorded
instant is the primary ordering and the pipeline state is the tiebreaker, so a replayed or
duplicated write can never silently overwrite a newer record with older field values.

## Carry the record forward on every write

Because a write renders the whole record and omission clears a field (**The record** above), every
command must first `flow state get` the existing record and carry forward every field it does not
itself own — `artifactUrl`, `jiraIssue`, `prUrl` and `worktrees` among them — before calling `flow
state set`. Re-emit each as read (`null` only if it was already `null`). Dropping one erases it
permanently: the link to the Jira issue, the PR (which also silently
downgrades the next fix from commit-and-push to staged-only), or the authoritative list of worktrees
for a multi-repo change.

## A change spanning repositories is one record

A change may affect more than one repository, and the record already treats that as **one**
record: a single scalar `branch` plus the `worktrees` map above, one entry per affected repository.
The store reads the same way — a two-repository change is one record, never two.

The **project key names the project whose state directory owns the record** — it is not the list of
affected repositories. The daemon derives the affected-repository set from `worktrees` on every
write and persists it alongside the record in the same transaction, so no reader ever observes a
change with a partially updated repository set. A skill writes `worktrees` and never supplies a
repository set separately.

The **key set of `worktrees` is the authoritative recorded list of affected worktrees** — it is
what `/flow`'s archive phase cleans up, and what resolves an app's root when a handoff needs an absolute
path. It is the record, not the iteration set: a step that needs "the worktrees" resolves that set
first, per **Resolving a change's worktrees** (`skills/flow-contracts/worktree-resolution.md`), rather than
looping over this map directly. The scalar `branch` names the shared branch only. Every repository
keeps its worktrees under `<project>/.worktrees/` — enforced by `check-worktree-location.sh`.

The **order** those repositories land in is not recorded here at all: it lives in the canonical
`link.md`'s `## Merge order`, which **Finish contract** (`skills/flow-contracts/finish-contract-run1.md`)
reads to sequence run 1's routes. The run itself is what keeps that record whole: every
repository of the resolved set, linked or not, is named in it at `flow.isolate-workspace`
(`skills/flow/implement.md`), so a repository no `spectre link` ever touched still has a
mechanical landing position instead of one reasoned out from prose mid-archive.

## Read it, write it

```bash
flow state get "$NAME" -C "$DIR"                    # prints the record, or falls back — see above
printf '%s' "$RECORD_JSON" | flow state set "$NAME" -C "$DIR"
```

`state set` reads the whole object from stdin — always write every field, never a partial merge, per
**The record** and **Carry the record forward on every write** above. Never write the on-disk
fallback file or journal directly; they belong to the CLI, are machine-local, and are **never
committed, never staged, and never archived** — nothing here is part of the change.
