# State self-heal

**Artifacts are the source of truth; the state file is a cache.** Every command reads the file first, then validates it against on-disk artifacts:

**Read artifacts from the apply worktree when one exists.** A change's real progress lives in its worktree, and `tasks.md` there is routinely ahead of the committed copy in the main checkout. Resolve the worktree from the state file's `worktrees` keys, or from `git worktree list` (branch `openspec/<name>`), and read `tasks.md`, `docs/manual-test/<name>.md`, and git state from **that** root. Fall back to the main checkout only when no worktree exists. Reading the main checkout while a worktree holds uncommitted work reports a change as far less advanced than it is — and can invite a `/myflow-do` re-run against real staged work.

| State claim | Contradicted when |
|-------------|-------------------|
| `STARTED` | an apply worktree for branch `openspec/<name>` exists |
| `IN_PROGRESS` | the change directory is already under `openspec/changes/archive/` |
| a recorded `prUrl` | PR existence was **conclusively** determined and no PR exists for the branch (a usable PR CLI for the host answered) — **the one permitted correction**, see below |
| a `worktrees` key | the path is absent from `git worktree list` |

**The one permitted correction.** Monotonicity forbids writing an earlier `state`, and this
carve-out does not: when a PR's non-existence is **conclusively** established (a usable PR CLI for
the host answered, and there is no PR), `prUrl` is set back to `null` while the state stays
`IN_PROGRESS`. It exists so an abandoned or closed-unmerged PR does not leave `/myflow-do`
committing and pushing to a branch nobody is reviewing, and so `/myflow-finish` offers the
integration choice again rather than waiting forever for a merge that cannot happen.

**A check that cannot be performed is not a contradiction.** If PR state cannot be determined (no PR CLI for the host, no network, no remote), treat it as **unknown** and leave `prUrl` as recorded. Never clear it on an inconclusive probe.

When the file is missing, unparseable, or contradicted: fall back to artifact inference, **rewrite the file with the inferred truth**, and announce the correction to the user (`⚠ state corrected: <old> → <new> (reason)`) before continuing. A hand-edited or out-of-band-modified file therefore degrades to artifact inference, never to a wrong answer.

Self-heal obeys the monotonicity rule stated under **State file** (`skills/myflow-contracts/state-file.md`): it may only raise or fill values. It infers `state` and — on a rebuild only, per the rule below — recovers the **keys** of `worktrees` from git, and never fabricates a `prUrl`.

**A rewrite carries forward every field it did not infer.**
**State file** (`skills/myflow-contracts/state-file.md`) requires every write to first read the
existing file and carry forward the fields it does not itself own; self-heal is a write, and the
fields it owns are `state` and — on a rebuild only, per the rule below — the **keys** of `worktrees`,
recovered from git rather than carried forward. `branch`, `artifactUrl`, `jiraIssue`,
`planningEffort`, `models` and `prUrl` are re-emitted exactly as read — never collapsed to `null` in
place of the value that was there — and so is `worktrees` itself on every path but that rebuild,
which is reached only when there was no successful read to carry it forward from. A write renders
the whole object, so a field left out of the render is not left unchanged, it is erased.

**For `planningEffort`, "exactly as read" is not a byte copy**, and this rewrite is bound by that
rule like every other write of a file it did not create: a file read through the retired-key
fallback is carried forward as its **mapped level under `planningEffort`**, per
**State file** (`skills/myflow-contracts/state-file.md`), which is canonical for it and is not
restated here. It applies to a file corrected for an unrelated contradiction exactly as it applies
to an ordinary write — re-emitting the old key leaves the file unmigrated, and writing
`planningEffort: null` because no such key was literally read erases a level the operator chose.

**When the prior file is missing or unparseable, most of those fields have no source.** Artifact
inference produces `state`; it has nothing to carry forward for the fields self-heal does not infer,
and writing `null` for them there is indistinguishable from a value that was legitimately never
set — the loss stays invisible until a later command can't find the proposal artifact or the tracker
issue. In that case, name every field that could not be recovered in the correction announcement
itself, so the loss is visible at the moment it happens rather than inferred later from a failure:

```text
⚠ state corrected: none → IN_PROGRESS (file unparseable; branch, artifactUrl, jiraIssue, planningEffort, models, prUrl unrecovered; worktrees recovered from this repository only, without merge bases — prior file could not be read)
```

**That example is exhaustive on purpose, and it is the template to copy.** A file that could not be
read leaves every field self-heal does not infer unrecovered, so each of them is named — including
`planningEffort` and `models`, which are exempt from this list only where a successful read showed
there was nothing to recover, per the two documented exceptions below. `worktrees` appears in a
clause of its own rather than among them because a rebuild recovers its keys and not its merge
bases, which is neither a clean carry-forward nor a loss; that is the next rule below.

**`worktrees` is the one field a rebuild recovers rather than nulls, because git already knows it.**
Scan the repository the command is running in for the worktrees on branch `openspec/<name>` with the
`git worktree list --porcelain` snippet under
**Worktree cleanup** (`skills/myflow-contracts/pipeline.md`) — the same mechanism `/myflow-finish`
already uses when the map is absent — and write one key per path found. An emptied map is not a
neutral loss: `/myflow-finish`'s preflight and its unfinished-work gate are both defined as *once
per recorded worktree*, so a map rebuilt to `{}` makes both **pass vacuously, having examined zero
worktrees**, for a change that may still hold an unmerged worktree with uncommitted work in it. A
key recovered from git is a fact, not an inference, so recovering it breaks no rule this contract
has: what it must never do is guess a path from a conventional layout, and this reads the path git
reports.

**The scan reaches one repository, and for a multi-repo change that is a real limit — named here
rather than left to be discovered.** A change spanning two repositories records both under
`worktrees`, and the multi-repo shape rule in **State file** (`skills/myflow-contracts/state-file.md`)
makes that key set the *authoritative* list of affected worktrees — which makes it the only on-disk
record of which repositories participate, and it is precisely the thing a rebuild could not read.
Nothing else names
them: `branch` is a scalar, and inferring a sibling path from a conventional layout is what this
contract forbids two paragraphs down. So a rebuild recovers the current repository's worktree and
**silently drops its siblings**, and the per-worktree gates in `/myflow-finish` then run over the
one that survived.

**That is a limit, not a regression, and the difference matters for what may be claimed.** Before
this rule a rebuild wrote `{}` and recovered nothing at all, so a sibling was lost either way; what
changed is that the current repository's worktree is now recovered. It is stated because the rule is
new and "recovered from git" would otherwise read as "all of them". Closing it would need a record
of the repository set that survives the file being unreadable, and this contract adds none — a fix
that cannot be made is not one to imply.

**What the scan cannot recover is the merge base**, which is a value only the run that created the
worktree ever knew — nothing on disk records it, and re-deriving one from today's base branch would
manufacture the exact number the preflight exists to compare against. So a recovered entry carries
`null` in its place, which **State file** (`skills/myflow-contracts/state-file.md`) defines as *no
recorded merge base*; every rule already written for a missing one governs it unchanged, and none of
them is restated here. The outcome is an honest unknown that makes the next `/myflow-finish` stop
and ask — which is what an empty map would have skipped past in silence. Name `worktrees` in the
announcement as *recovered from this repository only, without merge bases* whenever this happens, so
the operator learns both limits before reaching the refusal rather than at it — the missing merge
bases, and the sibling repositories a rebuild could not know to look in.

JSON that parses but is missing one or more of the fields this contract requires is treated as
unparseable in full, not partially recovered — the announcement above still names every field that
could not be read, exactly as if the file had failed to parse at all. A parseable-but-incomplete
file is not a smaller version of the all-or-nothing rule; it is the same case.

A file that read successfully but was merely *contradicted* by artifacts (the table above) carries
every unowned field forward from that successful read, so nothing is unrecovered in that case — the
announcement names only the state change, per the existing `⚠ state corrected: <old> → <new>
(reason)` shape.

**There is no legacy-value migration for the retired `stage` field.** A state file predating the
three-state model carries a `stage` field this contract does not recognise, which makes it
*unparseable* by the rule above — so it is rewritten from artifacts like any other unparseable file,
and the correction is announced. Backward compatibility was explicitly waived for that rename, and a
mapping table would be a second, drifting definition of a vocabulary that no longer exists.

**That claim is scoped to `stage`, and the retired-key exception below does not contradict it.** The
two cases differ in the one respect the argument turns on: `stage`'s twelve values were *collapsed*
into three that do not correspond to them, so there is no target to map onto and any table would be
inventing one. The planning effort's three levels were *renamed* one-to-one, every retired value
still having exactly one live level to name, so its mapping is a rename table over a live
vocabulary rather than a resurrection of a dead one, and it can be stated once without drifting
because both sides of it are still in force. Where a mapping would have to guess, this contract
still refuses one; that is what the paragraph above says, and it is why the exception below carries
the mapping's own boundary with it rather than inventing a level for a value outside it.

The schema is closed in both directions: just as a file missing a required field is unparseable,
**a file carrying any key not among this contract's documented fields is unparseable too** — that
is the general, mechanical rule for recognising a legacy field, without this file having to name
which fields have been retired.

**Two documented exceptions exist to the missing-field half of that rule: `planningEffort` and
`models`.** A file that omits either is valid and reads that field as *not recorded*, per
**State file** (`skills/myflow-contracts/state-file.md`). An omission the run **read** is therefore
not a failed recovery and is not named among the unrecovered fields — a file that did not read
successfully is the different case settled below; every other absent documented field still makes
the file unparseable.

**One documented exception exists to the undocumented-key half: the retired `effort` key.** A file
carrying it is read as recording the equivalent `planningEffort` level rather than as unparseable,
and is rewritten under the new key on the next write that file receives — without a correction being
announced, because the value was written correctly under the contract in force when it was set. The
level mapping is stated once, under
**Planning effort** (`skills/myflow-contracts/state-file.md`).

**The exception is unconditional on the value, and that is deliberate.** The mapping covers exactly
three values, but a file carrying anything else under that key is still *parseable*: the key is
excepted, and the unmapped value simply reads as *not recorded*. That rule, and the reason the
conditional form — parseable for a mapped value, unparseable for any other — was specified first and
withdrawn, are stated once with the `planningEffort` field under
**State file** (`skills/myflow-contracts/state-file.md`) and are not re-argued here: the argument
turns on what the *commands* do with such a file, which is that contract's subject, and two
independently authored copies of it had already drifted apart in wording. Its consequence for this
contract is one sentence — no file reaches self-heal on account of that key or its value, so there
is nothing here to do about either.

**Exemption from *being named among the unrecovered fields* is a second, narrower rule, and reading
the two as one is what makes an operator's level disappear.** That exemption holds where the
file **read successfully** and the run could therefore see that there was nothing to recover: the
key absent, or `effort` carrying a value outside the mapping. Nothing was lost in either case, so
naming it would report a loss that did not happen.

**A recorded level lost to a rebuild is named like every other field.** Every path that rebuilds is
a path where no read succeeded — the file missing, unparseable, parseable-but-incomplete, or
carrying an undocumented key — so the run saw neither the key nor its value, and a file that held a
real, operator-set level is rebuilt with that level nulled exactly as `artifactUrl` and `jiraIssue`
are. The exemption above does not reach those paths and must not be extended to them: doing so would
suppress the one announcement that makes the loss visible at the moment it happens, against this
contract's own rule that every field that could not be recovered is named. So `planningEffort` and
`models` are named in the unrecovered list on every rebuild, and the exemption applies only where a
successful read showed there was nothing there.

## The stale-`prUrl` gap — recorded, not closed

A state file whose `prUrl` is `null` while a PR actually exists is not detected by anything above,
because the table only checks the opposite direction. It arises when the file is hand-edited,
restored from a copy, or written by a session that then went further.

The consequence is real: `/myflow-finish` would treat run 1 as not yet done and offer to open a PR
for a branch that already has one, producing a second PR. Detecting it needs a PR probe, which is
exactly what is unavailable on the paths where this matters.

Before running `/myflow-finish` on a change you believe is already integrated, check
`git log @{upstream}..HEAD` and the forge. If the state file looks behind its branch, correct it
first.
