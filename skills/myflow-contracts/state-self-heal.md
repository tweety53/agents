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

Self-heal obeys the monotonicity rule stated under **State file** (`skills/myflow-contracts/state-file.md`): it may only raise or fill values. It infers `state` only, and never fabricates a `prUrl`.

**A rewrite carries forward every field it did not infer.** **State file**
(`skills/myflow-contracts/state-file.md`) requires every write to first read the existing file and
carry forward the fields it does not itself own; self-heal is a write, and its only owned field is
`state`. `branch`, `worktrees`, `artifactUrl`, `jiraIssue` and `prUrl` are re-emitted exactly as
read — never collapsed to `null` in place of the value that was there. A write renders the whole
object, so a field left out of the render is not left unchanged, it is erased.

**When the prior file is missing or unparseable, those fields have no source.** Artifact inference
only produces `state`; it has nothing to carry forward for the fields self-heal does not infer, and
writing `null` for them there is indistinguishable from a value that was legitimately never set —
the loss stays invisible until a later command can't find the proposal artifact or the tracker
issue. In that case, name every field that could not be recovered in the correction announcement
itself, so the loss is visible at the moment it happens rather than inferred later from a failure:

```text
⚠ state corrected: none → IN_PROGRESS (file unparseable; artifactUrl, jiraIssue unrecovered — prior file could not be read)
```

JSON that parses but is missing one or more of the fields this contract requires is treated as
unparseable in full, not partially recovered — the announcement above still names every field that
could not be read, exactly as if the file had failed to parse at all. A parseable-but-incomplete
file is not a smaller version of the all-or-nothing rule; it is the same case.

A file that read successfully but was merely *contradicted* by artifacts (the table above) carries
every unowned field forward from that successful read, so nothing is unrecovered in that case — the
announcement names only the state change, per the existing `⚠ state corrected: <old> → <new>
(reason)` shape.

**There is no legacy-value migration.** A state file predating the three-state model carries a
`stage` field this contract does not recognise, which makes it *unparseable* by the rule above —
so it is rewritten from artifacts like any other unparseable file, and the correction is
announced. Backward compatibility was explicitly waived for this rename, and a mapping table would
be a second, drifting definition of a vocabulary that no longer exists.

The schema is closed in both directions: just as a file missing a required field is unparseable,
**a file carrying any key not among this contract's documented fields is unparseable too** — that
is the general, mechanical rule for recognising a legacy field, without this file having to name
which fields have been retired.

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
