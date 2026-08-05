---
name: myflow-status
description: Show every open myflow change with its pipeline state, PR, next command, worktree, and last update. Read-only. Use for /myflow-status.
allowed-tools: Bash(openspec:*), Bash(git:*), Bash(jq:*), Bash(gh:*)
license: MIT
compatibility: Requires openspec CLI and jq.
metadata:
  author: gymie
  version: "2.0"
---

Report the pipeline state of every open (non-archived) OpenSpec change. **Read-only** — never commits, never runs git write operations, never advances a state. The one exception is state self-heal (see below), which corrects a stale cache to match the artifacts already on disk.

**Announce at start:** "Using myflow-status."

Follow all three contracts:

- **States** (`skills/myflow-contracts/pipeline.md`)
- **State file** (`skills/myflow-contracts/state-file.md`)
- **State self-heal** (`skills/myflow-contracts/state-self-heal.md`)

## Workflow

### 1. List open changes

Enumerate the candidate set exactly as **Change name resolution**
(`skills/myflow-contracts/pipeline.md`) defines it — the union of `openspec list --json` and the
names of the files in the project's state directory, minus anything archived — rather than running
`openspec list --json` alone: a change staged only in a worktree has a state file but no change
directory in this checkout, and would otherwise go unreported.

With a `<name>` argument, restrict to that change and include the detail view (step 4). With no argument, report every non-archived change.

Zero open changes → say so and suggest `/myflow-start`. Stop.

### 2. Resolve each change's state

For each change, resolve its user-scoped state file path per **State file** in `skills/myflow-contracts/state-file.md`, then read it:

```bash
MAIN_CHECKOUT="$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd)"
PROJECT_KEY="$(basename "$MAIN_CHECKOUT")-$(printf '%s' "$MAIN_CHECKOUT" | shasum | cut -c1-8)"
STATE_FILE="/Users/tweety53/Agents/myflow/state/$PROJECT_KEY/<name>.json"

jq -r '.state, .branch, .prUrl, .artifactUrl, .jiraIssue, (.planningEffort // .effort), (.models // {} | tojson), .updatedAt, .updatedBy, (.worktrees // {} | keys[])' \
  "$STATE_FILE" 2>/dev/null
```

**The planning-effort read falls back to the retired key, and that fallback is the whole of the
compatibility promise.** `(.planningEffort // .effort)` is what makes it real: jq's `//` yields the
left side whenever it is neither `null` nor `false`, so a file carrying only the retired key reads
its recorded value instead of reporting *not recorded*, and a file carrying both is governed by the
precedence **State file** (`skills/myflow-contracts/state-file.md`) states — the current key wins.
Reading `.planningEffort` alone would report a file that recorded a real level as having recorded
none, which is the one outcome the retired-key exception exists to prevent. The value that comes
back is the raw one; mapping it to a level is section 3's job, below.

Then validate against artifacts per **State self-heal** (`skills/myflow-contracts/state-self-heal.md`) — including its "read artifacts from the apply worktree when one exists" rule. Resolve the worktree **first**, then root every subsequent artifact check there:

1. worktree resolution — take the `worktrees` keys, or find the apply worktree for branch `openspec/<name>` in `git worktree list`; if found, treat its path as the artifact root for the rest of this step, otherwise use the main checkout
2. `tasks.md` (at the resolved root) — count `- [x]` vs `- [ ]` items
3. `docs/manual-test/<name>.md` (at the resolved root) — exists? any unchecked boxes?
4. merge status — decides whether the next `/myflow-finish` integrates or archives. It is answered
   **once per `worktrees` key**, in **three steps, in this order**:
   1. **Resolve the merge base recorded for that worktree** in the state file's `worktrees` map.
      **No recorded merge base**, and equally **one that is recorded but does not resolve in that
      worktree** — `git rev-parse --verify --end-of-options "<recorded>^{commit}"` fails, because
      history was rewritten, the clone is shallow, or the object was pruned — makes the merge status
      **inconclusive**. Do not infer one, and do not run either test below.
   2. **`HEAD` against that resolved merge base.** Equal → the branch has **no commits of its own**,
      so it is **not merged**, and no ancestor test is run.
   3. otherwise `git merge-base --is-ancestor <branch> origin/<base>`; a git failure or an
      unresolvable base ref is **inconclusive**, never "not merged"
5. PR — `gh pr list --head <branch> --state all --json number,state,url`, **only when `gh` is installed and the remote is a GitHub host**. Otherwise (Bitbucket, no `gh`, no network, branch never pushed) PR state is **unknown** — report it as unknown and treat it as inconclusive, never as "no PR exists".

**Steps 1 and 2 are not optional, and neither may be skipped because the ancestor test looks
decisive.** Why an unresolved or unequal-to-`HEAD` merge base has to be settled *before*
`git merge-base --is-ancestor` runs, and what answering them the other way round reports, is stated
once under **The block each state renders** (`skills/myflow-contracts/pipeline.md`), beside the
selection table this answer feeds. Read it there — it is not re-derived here, and it is not
re-derived from `scripts/check-finish-preflight.sh` either: that script's resolve-first guard and
its comment (b) are what `pipeline.md` cites in turn.

**A multi-repo change has one merge status per worktree, and the change's is the weakest of them.**
`worktrees` may carry more than one key, so item 4 is answered once per key and the answers must
then be combined. Combine them exactly as
**Finish contract** (`skills/myflow-contracts/pipeline.md`) already combines the preflight's
verdicts — run 2 proceeds only when **every** recorded worktree says so — which here means: the
change reads **merged** only when every worktree is proven merged; any worktree proven **not
merged** makes the change not merged; otherwise, with at least one inconclusive and none proven not
merged, the change is
**inconclusive**. Never report the first worktree's answer as the change's. When two worktrees
disagree, say so in the detail view and name which is which: that disagreement is precisely what the
next `/myflow-finish` will stop on, and the operator should see it here rather than discover it
there.

If the file is missing, unparseable, or contradicted, infer the state from artifacts, **rewrite the state file** — carrying forward every field this command did not infer, per **State self-heal** (`skills/myflow-contracts/state-self-heal.md`) — and mark the row `⚠`.

**Self-heal is monotonic, and writes almost nothing but `state`** (per **State file** in `skills/myflow-contracts/state-file.md`):

- Infer and correct **`state` only**, and **never fabricate a `prUrl`** — never invent one from a
  PR you happened to find, because a URL written from a guess is worse than a null.
- The **one** field beyond `state` a rebuild fills is `worktrees`, and only its **keys**, and only
  when the prior file could not be read at all — recovered from `git worktree list`, which reports a
  fact rather than an inference. The rule, what it cannot recover, and what the announcement must
  then say are **State self-heal** (`skills/myflow-contracts/state-self-heal.md`)'s. A file that read
  successfully is untouched here: its `worktrees` is carried forward exactly as read.
- **Do apply the one permitted correction** the contract defines: when this command's own
  `gh pr list` probe (step 2, item 5) **conclusively** answers that no PR exists for the branch
  while the state file records a `prUrl`, clear `prUrl` to `null` and mark the row `⚠`. This
  command already gathers exactly that evidence for its own report, so it is the one place the
  correction can be made without an extra probe. An inconclusive answer — no `gh`, non-GitHub
  forge, no network — clears nothing.
- **Never write a state earlier than the one recorded.** A check that cannot be performed is not a contradiction: an undeterminable PR state leaves `prUrl` exactly as recorded.

### 3. Render the table

Sort by state order per **States** in `skills/myflow-contracts/pipeline.md` (`STARTED`,
`IN_PROGRESS`, `FINISHED`), then by `updatedAt` descending. Omit `FINISHED` changes — they are
archived.

```
## myflow status

| Change | Jira | State | PR | Next | Worktree / branch | Updated |
|--------|------|-------|----|------|-------------------|---------|
| kan-8-myflow-updates | KAN-8 | IN_PROGRESS | #42 open | review the diff + run the guide, then `/myflow-finish` | `/abs/path/openspec-kan-8-myflow-updates` @ `openspec/kan-8-myflow-updates` | 2h ago (/myflow-do) |
| active-workout-session-editing | — | STARTED | — | read the artifact, then `/myflow-do` | none | 19h ago (/myflow-start) |

⚠ = state file was stale and has been corrected from artifacts.
```

Worktree paths are **absolute**, taken from the `worktrees` keys. The **PR** column shows the
number and state when known, `—` when `prUrl` is null, and `?` when PR state could not be
determined.

The **Jira** column shows `jiraIssue` verbatim, or `—` when the change has no linked issue. This
is a **read-only** report: never call Jira, never transition an issue, never infer a key from the
change name.

Surface `artifactUrl` when present — the link to the published proposal artifact.

Surface `planningEffort` the same way: the recorded level verbatim when there is one, and, when
step 2's read yielded nothing, `not recorded — planned at default`, that being the level recommended
under **Planning effort** in State file (`skills/myflow-contracts/state-file.md`) — that file is
canonical for the levels and for which of them is recommended, so read the set there rather than
inferring it from this line.

**A value that came from the retired key is surfaced as the level it maps to, never as the raw
value.** Step 2's read yields whichever key held a value; `medium` is surfaced as `default`, `high`
as `detailed` and `low` as `low`, per the mapping stated once under
**Planning effort** in State file (`skills/myflow-contracts/state-file.md`). Say which level is in
force and nothing more — the key it was read from is not the operator's business, and this report
never rewrites the file to migrate it. A raw `medium` printed as a level would name a level this
pipeline does not have.

**A retired-key value outside those three maps to no level, and is surfaced as `not recorded —
planned at default` rather than echoed.** It is not a `⚠` and does not make the file unparseable:
the mapping's boundary is the level's, not the schema's, and what an unmapped value reads as is
stated once under **Planning effort** in State file (`skills/myflow-contracts/state-file.md`), where
the reason it is *not recorded* rather than *unparseable* is recorded too. This line reports the
level, and there is none.

An absent `planningEffort` is legal and is never a `⚠`:
per **State file** (`skills/myflow-contracts/state-file.md`), a file that omits it reads as `null`,
so it is neither a contradiction nor a field to infer. This report never writes `planningEffort`; it
carries it forward untouched on a self-heal, like every other field it does not own.

Surface `models` the same way, as one line covering its three roles — `implementation`,
`reviewPanel` and `panelFix` — each the recorded model verbatim, or `not recorded` where none was
chosen. The same rules hold: an absent `models` object is legal and is never a `⚠`, per
**State file** (`skills/myflow-contracts/state-file.md`); and this report never writes the field,
carrying it forward untouched on a self-heal.

Next-command mapping:

| State | Next |
|-------|------|
| `STARTED` | read the artifact, then `/myflow-do <name>` (or re-run `/myflow-start` to revise) |
| `IN_PROGRESS`, branch not merged | review the diff + run the guide, then `/myflow-finish <name>` (or re-run `/myflow-do` to fix) |
| `IN_PROGRESS`, branch merged | `/myflow-finish <name>` — it will archive |
| `FINISHED` | — |

The `IN_PROGRESS` row splits on merge status because `/myflow-finish` behaves differently either
side of it: it integrates before the merge and archives after. Say which run the operator is
about to get. **Merge status here is step 2 item 4's answer — all three of its steps, and combined
across every worktree** — never a bare ancestor test. A branch with no commits of its own is *not
merged*, and an inconclusive answer takes the not-merged row and says the check could not be
completed.

### 4. Detail view (only when `<name>` was given)

Add below the table:

- Linked Jira issue key, or "none linked"
- Task progress (`N/M` checked from `tasks.md`)
- Nested `<name>-fix-N` sub-changes, if any
- The manual test guide's **absolute** path + checked/total box count
- PR number, state, and URL when one exists
- Whether the branch has reached the base branch — i.e. which `/myflow-finish` run comes next. For a
  change recording more than one worktree, give the combined answer **and** the per-worktree ones
  whenever they differ, naming each worktree by its absolute path
- Every `⚠` correction made, with the reason

Then regenerate the change's full handoff block for its current state and print it, rendered from
the per-state template in **Handoff output** (`skills/myflow-contracts/pipeline.md`). Build it from
the state file and the artifacts as they now stand; nothing is read back from a stored copy of an
earlier run's text, because no command stores one.

- A value the state file does not carry is reported as **missing**, so a block whose artifact URL
  reads *missing* is distinguishable from one whose URL was never printed. A run-only value is the
  exception and is omitted instead:
  **The block each state renders** (`skills/myflow-contracts/pipeline.md`) marks which fields those
  are, and this file does not list them again.
- `IN_PROGRESS` renders two different blocks, and **the merge status this command already computed
  in step 2 chooses between them whenever it is conclusive** — a branch proven to have reached the
  base branch is integrated, so the branch-waiting-on-a-merge block is the right one for it, and its
  heading says the merge has happened. `prUrl` is consulted only where that probe was inconclusive.
  The selection table, and the reason the weaker signal must not run ahead of the stronger one, are
  under **The block each state renders** (`skills/myflow-contracts/pipeline.md`); render from that
  table rather than from a rule restated here.
- **Use the answer you already have, and never re-derive it.** Step 2 item 4 already produced the
  merge status — all three of its steps, combined across every worktree — for the next-command
  column, and re-deriving it here with a bare ancestor test would reintroduce the
  no-commits-of-its-own false positive that item names.
  Reading `prUrl` in front of that answer is what let one invocation
  report *branch merged → it will archive* in the table and *waiting on the merge* in the block, for
  the same change, in the same run — a change stopped at a run-2 cleanup leftover is exactly that
  case, and it is not rare.
- **The two splits still do not compete.** The next-command column splits `IN_PROGRESS` on merge
  status to say which `/myflow-finish` run the operator gets; the block splits on it to say which
  wait the operator is in. Both now read the same signal first, so they cannot disagree about the
  branch — and because both blocks end in `/myflow-finish <name>`, neither can contradict the other
  about what to run next.
- The `prUrl` test that remains is one-way — a `null` `prUrl` does not prove run 1 has not
  happened — and what that costs, plus why it is accepted rather than replaced, is stated under
  **The block each state renders** (`skills/myflow-contracts/pipeline.md`). Do not restate that
  reasoning here, and do not present the test as conclusive.
- `FINISHED` changes have no regenerated block, exactly as they have no row.

With no change name there is no block at all: the no-argument report stays the table above,
unchanged.

## Guardrails

- **Never** commit, stage, push, merge, or archive.
- **Never** advance a state — only correct a stale cache to match artifacts.
- **Never** rewind a state, and never fabricate a `prUrl`. Clearing a `prUrl` that a conclusive
  probe disproved is the one permitted correction, per
  **State self-heal** (`skills/myflow-contracts/state-self-heal.md`); everything else about the
  file is read-only.
- **Never** create a worktree or branch.
- Report `gh` being unavailable, or a non-GitHub forge, as "PR state unknown" rather than guessing — and never clear `prUrl` on that unknown.
- Never guess a state when artifacts are ambiguous — show `?` and say which check was inconclusive.
- **Never** act on a regenerated handoff block — it is output, not a plan. No command it names is
  executed, nothing is staged, and no state is written on account of it, however imperative its last
  line reads.
- **No flags.** The only argument is the optional change name.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Status of all open changes | `/myflow-status` |
| Detail for one change | `/myflow-status <name>` |
| Pipeline reference | `/myflow-info` |
