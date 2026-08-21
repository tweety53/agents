---
name: myflow-status
description: Show every open myflow change with its pipeline state, PR, next command, and last update. Read-only. Use for /myflow-status.
allowed-tools: Bash(openspec:*), Bash(git:*), Bash(jq:*), Bash(myflow:*)
license: MIT
compatibility: Requires openspec CLI, the myflow CLI, and jq.
metadata:
  author: gymie
  version: "2.0"
---

Report the pipeline state of every open (non-archived) OpenSpec change. **Read-only** — never commits, never runs git write operations, never advances a state, and never writes state.

**Announce at start:** "Using myflow-status."

Follow both contracts:

- **States** (`skills/myflow-contracts/pipeline.md`)
- **State file** (`skills/myflow-contracts/state-file.md`)

## Workflow

### 1. List open changes

**Check guard presence.** Per **Guard presence check** (`skills/myflow-contracts/pipeline.md`),
confirm the one guard this command invokes — `resolve-base-branch.sh` — is present in
`skills/myflow-status/scripts/`. A complete set prints nothing; an absence prints that section's
block once, and the run continues under the guard's own hand-run fallback. Step 2 below still
reimplements `check-finish-preflight.sh`'s other merge-status steps in prose rather than invoking
that script (see the note there); only base-branch resolution is delegated to a real guard.

Enumerate the candidate set exactly as **Change name resolution**
(`skills/myflow-contracts/pipeline.md`) defines it — through `myflow state list [-C dir]`, never a
hand-written HTTP call:

```bash
MAIN_CHECKOUT="$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd -P)"
BOARD="$(myflow state list -C "$MAIN_CHECKOUT")"
```

`$BOARD` is one JSON object: `"source"` (`"store"` or `"fallback"`), `"complete"`
(`true` only when `"source":"store"`), and `"records"` (each `name`, `state`, `updatedAt`,
`updatedBy`, and `"unreadable":true` for a fallback file that could not be parsed).

**Report which source produced the set, before the table** — `jq -r .source <<<"$BOARD"`. When it
reads `fallback`, say so in one line — `⚠ store unreachable — reporting from local fallback files,
which may be stale` — so a report built during an outage is never mistaken for one built from the
live store. Say nothing extra when it reads `store`; that is the normal case.

With a `<name>` argument, restrict to that change and include the detail view (step 4). With no argument, report every non-archived change.

Zero open changes → say so and suggest `/myflow-start`. Stop.

### 2. Resolve each change's state

For each change, resolve its record through the CLI, per **State file**
(`skills/myflow-contracts/state-file.md`) — never a direct file read, never `jq` on a path:

```bash
MAIN_CHECKOUT="$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd -P)"
ERR="$(mktemp)"
RECORD="$(myflow state get "<name>" -C "$MAIN_CHECKOUT" 2>"$ERR")"
STATUS=$?
WARNING="$(cat "$ERR")"; rm -f "$ERR"
```

- **`STATUS=0` and `$WARNING` empty** — the record came from the store.
- **`STATUS=0` and `$WARNING` reads `⚠ myflow: store unreachable — read local fallback`** — the
  record came from the on-disk fallback file. Report this change's row as **source: fallback**. If
  `$RECORD` is also empty (no fallback file exists either), treat this exactly like `STATUS=1`
  below.
- **`STATUS=1`** — the store was reached and correctly holds no record for this change (`myflow:
  no state recorded for <project>/<name>` on stderr): report the change by name as **no state
  recorded** and omit it from the table, the same way a missing state file was always reported.
- **Any other `STATUS`, or a `$RECORD` that is not valid JSON** — the record is unreadable: name the
  change in this command's own output and skip it. Never rebuild it by inference.

Then read the fields from `$RECORD`:

```bash
printf '%s' "$RECORD" | jq -r '.state, .branch, .prUrl, .artifactUrl, .jiraIssue, (.planningEffort // .effort), (.models // {} | tojson), (.reviewPanelRoster // null), .updatedAt, .updatedBy, (.worktrees // {} | keys[])'
```

**The planning-effort read falls back to the retired key, and that fallback is the whole of the
compatibility promise.** `(.planningEffort // .effort)` is what makes it real: jq's `//` yields the
left side whenever it is neither `null` nor `false`, so a file carrying only the retired key reads
its recorded value instead of reporting *not recorded*, and a file carrying both is governed by the
precedence **State file** (`skills/myflow-contracts/state-file.md`) states — the current key wins.
Reading `.planningEffort` alone would report a file that recorded a real level as having recorded
none, which is the one outcome the retired-key exception exists to prevent. The value that comes
back is the raw one; mapping it to a level is section 3's job, below.

**Merge status** decides whether the next `/myflow-finish` integrates or archives. It is answered
once per worktree in the set resolved per **Resolving a change's worktrees**
(`skills/myflow-contracts/pipeline.md`) — never a raw read of the record's `worktrees` map,
which a `{}` or absent map would make a loop over its keys report on nothing. Per that same
section, a resolved set that comes back empty is never a vacuous pass: say so in this change's
detail view — **merge status: unknown, no worktree recorded** — rather than silently omitting the
row. Each worktree in the resolved set is answered in **three steps, in this order**:

1. **Resolve the merge base recorded for that worktree** in the record's `worktrees` map.
   **No recorded merge base**, and equally **one that is recorded but does not resolve in that
   worktree** — `git rev-parse --verify --end-of-options "<recorded>^{commit}"` fails, because
   history was rewritten, the clone is shallow, or the object was pruned — makes the merge status
   **inconclusive**. Do not infer one, and do not run either test below.
2. **`HEAD` against that resolved merge base.** Equal → the branch has **no commits of its own**,
   so it is **not merged**, and no ancestor test is run.
3. otherwise resolve `<base>`, **in the same worktree**, by invoking `resolve-base-branch.sh`
   exactly as **Finish contract** (`skills/myflow-contracts/finish-contract.md`) does — never a
   hand-derived name, and never `HEAD@{upstream}`, for the reason stated there; running it in this
   worktree is also what satisfies its unconditional assertion that the base differs from the
   current branch, since `HEAD` here is the change's own branch. **A non-zero exit is
   inconclusive** — the same disposition this step already gives a git failure or an unresolvable
   base ref. The report continues, and the detail view carries the resolver's own stderr message
   rather than inventing one; `/myflow-status` is read-only and never blocks. On exit `0`, run
   `git merge-base --is-ancestor <branch> origin/<base>`; a git failure there is likewise
   **inconclusive**, never "not merged"

**Steps 1 and 2 are not optional, and neither may be skipped because the ancestor test looks
decisive.** Why an unresolved or unequal-to-`HEAD` merge base has to be settled *before*
`git merge-base --is-ancestor` runs, and what answering them the other way round reports, is stated
once under **The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`), beside the
selection table this answer feeds. Read it there — it is not re-derived here, and it is not
re-derived from `<agents repo>/scripts/check-finish-preflight.sh` either: that script's resolve-first guard and
its comment (b) are what `pipeline.md` cites in turn.

**A multi-repo change has one merge status per worktree, and the change's is the weakest of them.**
The resolved set may carry more than one worktree, so the merge-status check is answered once per
worktree in it and the answers must then be combined. Combine them exactly as
**Finish contract** (`skills/myflow-contracts/finish-contract.md`) already combines the preflight's
verdicts — run 2 proceeds only when **every** worktree in the resolved set says so — which here
means: the
change reads **merged** only when every worktree is proven merged; any worktree proven **not
merged** makes the change not merged; otherwise, with at least one inconclusive and none proven not
merged, the change is
**inconclusive**. Never report the first worktree's answer as the change's. When two worktrees
disagree, say so in the detail view and name which is which: that disagreement is precisely what the
next `/myflow-finish` will stop on, and the operator should see it here rather than discover it
there.

### 3. Render the table

Sort by state order per **States** in `skills/myflow-contracts/pipeline.md` (`STARTED`,
`IN_PROGRESS`, `FINISHED`), then by `updatedAt` descending. Omit `FINISHED` changes — they are
archived.

```
## myflow status

| Change | Jira | State | PR | Next | Updated |
|--------|------|-------|----|------|---------|
| kan-8-myflow-updates | KAN-8 | IN_PROGRESS | #42 | review the diff + run the apps, then `/myflow-finish` | 2h ago (/myflow-do) |
| active-workout-session-editing | — | STARTED | — | read the artifact, then `/myflow-do` | 19h ago (/myflow-start) |
```

The absolute worktree path is given in the detail view, taken from the `worktrees` keys.

The **PR** column shows the number parsed from the recorded `prUrl`, or `—` when it is `null`. It
never reports whether the pull request is open, merged or closed — that answer needs a network call
this command no longer makes; the detail view says where to look instead.

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
planned at default` rather than echoed.** It does not make the file unparseable:
the mapping's boundary is the level's, not the schema's, and what an unmapped value reads as is
stated once under **Planning effort** in State file (`skills/myflow-contracts/state-file.md`), where
the reason it is *not recorded* rather than *unparseable* is recorded too. This line reports the
level, and there is none.

Surface `models` the same way, as one line covering its three roles — `implementation`,
`reviewPanel` and `panelFix` — each the recorded model verbatim, or `not recorded` where none was
chosen.

Surface `reviewPanelRoster` the same way: the recorded preset verbatim when there is one, and,
when step 2's read yielded `null`, `not recorded — using the default`, that default being `light`
per `skills/myflow-start/SKILL.md`'s roster prompt, where it is the recommended option; what each
preset means is canonical in `skills/myflow-do/SKILL.md`. A
not-recorded roster is not a warning: it is the default, and this line reports it as a normal
state, not as something missing.

Next-command mapping:

| State | Next |
|-------|------|
| `STARTED` | read the artifact, then `/myflow-do <name>` (or re-run `/myflow-start` to revise) |
| `IN_PROGRESS`, branch not merged | review the diff + run the apps, then `/myflow-finish <name>` (or re-run `/myflow-do` to fix) |
| `IN_PROGRESS`, branch merged | `/myflow-finish <name>` — it will archive |
| `FINISHED` | — |

The `IN_PROGRESS` row splits on merge status because `/myflow-finish` behaves differently either
side of it: it integrates before the merge and archives after. Say which run the operator is
about to get. **Merge status here is step 2's answer — all three of its steps, and combined
across every worktree** — never a bare ancestor test. A branch with no commits of its own is *not
merged*, and an inconclusive answer takes the not-merged row and says the check could not be
completed.

### 4. Detail view (only when `<name>` was given)

Add below the table:

- Linked Jira issue key, or "none linked"
- Task progress (`N` of `M` checked, from `tasks.md`)
- Nested `<name>-fix-N` sub-changes, if any
- PR number and URL when one exists — not whether it is open, merged or closed, which this report
  does not track; check the forge for that
- Whether the branch has reached the base branch — i.e. which `/myflow-finish` run comes next. For a
  change recording more than one worktree, give the combined answer **and** the per-worktree ones
  whenever they differ, naming each worktree by its absolute path

Then regenerate the change's full handoff block for its current state and print it. Load **Handoff
blocks** (`skills/myflow-contracts/handoff-blocks.md`) here — it is canonical for the per-state
templates, and this is the only step of any `/myflow-*` command that loads it — and render from the
template that matches the current state. Build it from the record and the artifacts as they now
stand; nothing is read back from a stored copy of an earlier run's text, because no command stores
one.

- A value the record does not carry is reported as **missing**, so a block whose artifact URL
  reads *missing* is distinguishable from one whose URL was never printed. A run-only value is the
  exception and is omitted instead:
  **The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`) marks which fields those
  are, and this file does not list them again.
- `IN_PROGRESS` renders two different blocks, and **the merge status this command already computed
  in step 2 chooses between them whenever it is conclusive** — a branch proven to have reached the
  base branch is integrated, so the branch-waiting-on-a-merge block is the right one for it, and its
  heading says the merge has happened. `prUrl` is consulted only where that probe was inconclusive.
  The selection table, and the reason the weaker signal must not run ahead of the stronger one, are
  under **The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`); render from that
  table.
- **Use the answer you already have, and never re-derive it.** Step 2 already produced the
  merge status — all three of its steps, combined across every worktree — for the next-command
  column, and re-deriving it here with a bare ancestor test would reintroduce the
  no-commits-of-its-own false positive that check names.
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
  **The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`). Do not restate that
  reasoning here, and do not present the test as conclusive.
- `FINISHED` changes have no regenerated block, exactly as they have no row.
- **The `Run it:` section is resolved, never copied from a stored run.** Follow **6. Resolve the
  run instructions** (`skills/myflow-do/SKILL.md`) — canonical for how those lines are produced —
  and apply it here exactly as `/myflow-do` does: resolve from the worktree named in the record
  and `<project>/.myflow/project.md` — never the project's declared base — not from any text
  `/myflow-do` printed earlier. Do not restate the resolution *procedure* here — the steps that
  compute each app root, start command and URL; a second copy of those steps is the failure this
  repository's contracts are built to avoid, and naming the invariant above is not one.

With no change name there is no block at all: the no-argument report stays the table above,
unchanged.

## Guardrails

- **Never** commit, stage, push, merge, or archive.
- **Never** advance a state, rewind a state, or write the record — this command is entirely
  read-only.
- **Never** fabricate a `prUrl`.
- **Never** create a worktree or branch.
- **Never** call `gh`, or any other network command, to determine PR state — the PR column reports
  only the number parsed from the recorded `prUrl`.
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
