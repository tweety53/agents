# State self-heal

**Artifacts are the source of truth; the state file is a cache.** Every command reads the file first, then validates it against on-disk artifacts:

**Read artifacts from the apply worktree when one exists.** A change's real progress lives in its worktree, and `tasks.md` there is routinely ahead of the committed copy in the main checkout. Resolve the worktree from `git worktree list` (branch `openspec/<name>`) and read `tasks.md`, `docs/manual-test/<name>.md`, and git state from **that** root. Fall back to the main checkout only when no worktree exists. Reading the main checkout while a worktree holds uncommitted work reports a change as far less advanced than it is — and can invite a `/myflow-do` re-run against real staged work.

| State claim | Contradicted when |
|-------------|-------------------|
| `stage: awaiting-proposal-review` / `proposal-done` | an apply worktree for branch `openspec/<name>` exists; or, when no worktree exists, every task in the main checkout's `tasks.md` is checked |
| `stage: awaiting-do-review`, `do-review-started`, `do-done` | `docs/manual-test/<name>.md` exists and is newer than the last code change |
| `stage: awaiting-manual-test`, `manual-test-done` | the branch has commits beyond `MERGE_BASE` **and** a PR is confirmed to exist for it |
| `stage: awaiting-pr-review` | PR existence was **conclusively** determined and no PR exists for the branch (a usable PR CLI for the host answered) — **the one permitted rewind**, to `manual-test-done`, see below |
| `worktree` path | the path is absent from `git worktree list` |

**The one permitted rewind.** Monotonicity forbids writing an earlier `stage`, with a single
carve-out: when a PR's non-existence is **conclusively** established (a usable PR CLI for the host
answered, and there is no PR), a change at `awaiting-pr-review` may be rewound to
**`manual-test-done`** and `gates.prOpened` set back to `false`. This is the only backward stage
write of that kind in the system, and it exists so an abandoned or closed-unmerged PR does not
strand a change forever. Gate values other than `prOpened` are still sticky — never lower
`gates.tested`. If the PR check is inconclusive, do not rewind (see below).

**The rewind target is `manual-test-done`, not `awaiting-manual-test`.** Testing is finished and
recorded (`gates.tested` is `true` or `"skipped"`); only the PR went away. The user's next action
is `/myflow-review <name>` to open a fresh one — not re-running the manual test. Rewinding to
`awaiting-manual-test` would also leave a `true`/`"skipped"` gate coexisting with a stage whose
whole meaning is "testing still pending", which is exactly the inconsistency monotonic gates exist
to prevent.

**A check that cannot be performed is not a contradiction.** If PR state cannot be determined (no PR CLI for the host, no network, no remote), treat it as **unknown** and leave the stage as recorded. Never rewind a stage on an inconclusive probe.

When the file is missing, unparseable, or contradicted: fall back to artifact inference, **rewrite the file with the inferred truth**, and announce the correction to the user (`⚠ state corrected: <old> → <new> (reason)`) before continuing. A hand-edited or out-of-band-modified file therefore degrades to the old artifact-inference behavior, never to a wrong answer.

Self-heal obeys the monotonic-gate rule stated under **State file** (`skills/myflow-contracts/state-file.md`): it may only raise or fill values. It infers `stage` only, preserves existing gate values as they stand, fills `null` gates conservatively (`false`, never `true`), and never writes `gates.tested: true`.

**Legacy stage values.** A state file written before the 12-stage model may carry `start`,
`awaiting-review`, or `awaiting-test`. Map them on read — `start` → `proposal-done`, <!-- vocab-guard:allow -->
`awaiting-review` → `awaiting-do-review`, `awaiting-test` → `awaiting-manual-test` — rewrite the <!-- vocab-guard:allow -->
file, and announce the correction. `awaiting-pr-review` and `finished` are unchanged.

## Narrowing on the seven pure-state-write commands' happy path

`state-advance.sh` (`skills/myflow-state-advance/state-advance.sh`) performs a **strict subset** of
this contract's checks, and only for the seven `*-done` / `*-manual-review` commands:

- **What it checks:** the state file exists and parses as a JSON **object** carrying the structural
  keys — `stage`, and `gates` with all four gate values; the current `stage` is one of the values
  passed via `--accepted`; and, when the state file names a non-null `worktree`, that the
  repository owning that path still lists it. A field this contract allows to be absent rather than
  `null` (`fastPath`) is **not** required — only the structural keys are, so that a
  `{"stage": …}`-shaped fragment escalates while a legitimate file with an omitted optional field
  does not.
- **What it does not check:** every artifact-contradiction row in the table above — `tasks.md`
  completion, `docs/manual-test/<name>.md` presence/recency, commits beyond `MERGE_BASE`, or PR
  existence/non-existence (including the one permitted rewind). It also does not verify that a
  worktree path outside any repository git can answer for is stale — an unanswerable probe is
  treated as unknown, per **A check that cannot be performed is not a contradiction** above. None
  of that runs on the script's happy path.
- **Scope:** this narrowing applies **only** to the script's happy path (exit `0`) for these seven
  commands. Any escalation (exit `3`, `4`, `5`, or `6`) hands off to the `myflow-state-advance`
  skill, which applies this contract in full, unchanged.

**One gap this leaves open — recorded, not closed.** The dropped row "`stage: awaiting-manual-test`,
`manual-test-done` contradicted when the branch has commits beyond `MERGE_BASE` **and** a PR is
confirmed to exist" was the only check anywhere that detected a **backward-stale** state file: one
whose recorded stage is behind where the branch actually is, because the file was hand-edited,
restored from a copy, or written by a session that then went further. Nothing downstream catches
that. `/myflow-review` treats `manual-test-done` as its normal entry stage, so it would push and
open a **second PR** for a branch that already has one.

The script deliberately does **not** add a stage-monotonicity check to compensate. Its
`--target originStage` mode legitimately writes an *earlier* stage — that is the whole fix-re-entry
mechanism — so a naive "never write a stage before the current one" test would break every fix
round. Until a check exists that can tell a legitimate fix re-entry from a stale file, this is a
known hole: if a state file looks behind the branch, correct it through this skill before running
`/myflow-review`.

**Why the rest of this narrowing is acceptable.** The seven stages these commands write are pure human-confirmation
writes — a person stating "I reviewed this" or "I tested this" — not claims about code or PR state
that need artifact corroboration at the moment they're recorded. The commands that actually gate on
artifact truth, `/myflow-review` and `/myflow-finish`, independently verify Gate C coverage, test/
lint results, and PR merge state before they act — so a narrower check here does not let a false
claim reach an irreversible action uncaught.
