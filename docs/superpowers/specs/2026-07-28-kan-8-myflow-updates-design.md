# KAN-8 — myflow updates: five states, six commands

**Jira:** KAN-8 "Myflow updates"
**Date:** 2026-07-28
**Repo:** `/Users/tweety53/Projects/agents`

## Why

myflow grew to twelve stages, eighteen commands and nineteen skills. Seven of those commands
are pure state writes that exist only to record "a human looked at this", and the stage
vocabulary encodes both *how far the work got* and *who is waiting* in a single field — which
is why it needed twelve values, an `originStage` field to remember where a fix came from, and
a monotonicity contract to keep gates from being demoted.

The cost lands in three places. The contract that describes all of it is large enough that
moving it out of the always-on rule layer was its own ticket (KAN-10, then
`myflow/lazy-load-pipeline`). Every rename has to be swept across two command trees, nineteen
skills, three root docs and a vocabulary guard. And the operator has to remember which of
eighteen commands is legal at the current stage.

KAN-8 collapses the model: **the state says how far the work got, and the human gate is a
property of that state.** Five states, six pipeline commands, and no command whose only job is
to say "done".

## What changes

### 1. The state machine

Five states. Each pipeline command ends in the state named after it.

```
              /myflow-start
        ─────────────────────►  STARTED       human: read the proposal artifact
                                   │
                                   │ /myflow-do
                                   ▼
                               IN_PROGRESS    human: review the staged diff
                                   │
                                   │ /myflow-test
                                   ▼
                                 TEST         human: run the apps
                                   │
                                   │ /myflow-review
                                   ▼
                                REVIEW        human: review + merge the PR
                                   │
                                   │ /myflow-finish
                                   ▼
                               FINISHED
```

`/myflow-fast` creates a change and lands directly at REVIEW.

**Every command is re-entrant.** `/myflow-start` re-run at STARTED revises the proposal and
republishes the artifact to the same URL — this replaces `/myflow-start-fix`. `/myflow-do`
re-run at IN_PROGRESS, TEST or REVIEW applies a fix — this replaces `/myflow-do-fix`.

**A fix never moves the state.** `/myflow-do` only ever advances STARTED → IN_PROGRESS; from
any later state it leaves the state exactly as it found it. This is what removes `originStage`
entirely: the state is a fact about how far the work got, not about where a fix was raised.

The consequence is deliberate and accepted: a fix applied at TEST does not re-open the test
gate, and a fix applied at REVIEW does not re-open the diff gate. The operator decides whether
to re-run an earlier command. myflow does not decide it for them.

**Skipping the manual test is not running `/myflow-test`.** `/myflow-review` accepts
IN_PROGRESS *or* TEST. Arriving from IN_PROGRESS means testing was skipped, and review records
`tested: "skipped"`. There is no skip prompt, no guide written, no checklist to parse, and no
sticky-gate rule to enforce — the absence of the command is the skip.

### 2. Command contracts

| Command | Accepts | Git it may do | Ends at |
|---------|---------|---------------|---------|
| `/myflow-start` | *(none)* or STARTED | none — planning artifacts only | STARTED |
| `/myflow-do` | STARTED, IN_PROGRESS, TEST, REVIEW | worktree + `git add`; **commit + push only from REVIEW** | IN_PROGRESS from STARTED; otherwise unchanged |
| `/myflow-test` | IN_PROGRESS | `git add` the guide | TEST |
| `/myflow-review` | IN_PROGRESS, TEST | commit, push, open PR; merge only if chosen at invocation | REVIEW |
| `/myflow-finish` | REVIEW | commit + push the archive; remove worktrees and branches | FINISHED |
| `/myflow-fast` | *(none)* | everything through opening the PR | REVIEW |
| `/myflow-info`, `/myflow-status` | any — read-only | none | unchanged |

The REVIEW carve-out for `/myflow-do` is the one place a fix commits: a PR already exists
remotely, so a staged-only fix would be invisible on it.

**On a state mismatch, stop.** Report the actual state, what the command expects, and which
command to run instead. Ask for an explicit override with "run the suggested command instead"
as the default. Never advance from a wrong starting state silently.

### 3. Auto-merge becomes a choice

The `automerge` flag is removed. `/myflow-review` asks **before doing any work** how the run
should end:

- **Open a PR and stop** (default) → REVIEW, human merges on the forge
- **Merge it when checks pass** → commits, pushes, merges → REVIEW, with nothing left for the
  human to do there but run `/myflow-finish`

Both answers end at REVIEW. The state records how far the work got; whether a human still owes
it a merge is answered by `prUrl` and the branch's merge status, which finish re-verifies
anyway. Asking first means the run then proceeds to completion without another interruption.
The answer is never remembered between runs and never inferred from another flag.

### 4. State file

```json
{
  "state": "TEST",
  "branch": "openspec/kan-8-myflow-updates",
  "worktrees": { "/abs/path/to/worktree": "<merge-base sha>" },
  "artifactUrl": "https://…",
  "jiraIssue": "KAN-8",
  "prUrl": null,
  "tested": null,
  "fast": false,
  "updatedAt": "2026-07-28T21:00:00Z",
  "updatedBy": "/myflow-test"
}
```

Path resolution is unchanged: `--git-common-dir` → `<project-key>` →
`/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json`. The file still lives outside
the repo and is never staged, committed or archived.

Removed fields, each with the reason it no longer has a job:

| Field | Why it goes |
|-------|-------------|
| `gates.reviewed` | Nothing records it once the `*-done` commands are gone |
| `gates.prOpened` | `prUrl` non-null says it |
| `gates.prMerged` | FINISHED says it, and finish verifies against git regardless |
| `originStage` | Fixes no longer move the state, so there is nothing to return to |
| `REVIEWED_TREE` | Only existed for the fast-path checkpoint stop, which is removed |
| `worktree` + `MERGE_BASE` | Merged into one `worktrees` map keyed by absolute path |

`tested` is `null` (not yet reviewed), `"skipped"` (review ran from IN_PROGRESS), or `true`
(review ran from TEST with the guide's boxes checked). `/myflow-review` is its only writer.

**Monotonicity** keeps one rule and one exception: never write an earlier state, except
REVIEW → TEST when a PR's absence is *conclusively* established (a usable PR CLI answered and
there is no PR). An inconclusive probe — no CLI, no network, no remote — is not a
contradiction and rewinds nothing.

**Self-heal** survives, shrunk to the states that still have artifacts able to contradict them:

| State claim | Contradicted when |
|-------------|-------------------|
| STARTED | a worktree for `openspec/<name>` exists |
| IN_PROGRESS | `docs/manual-test/<name>.md` exists and is newer than the last code change |
| TEST | the branch has commits beyond its merge base **and** a PR is confirmed to exist |
| REVIEW | PR non-existence was conclusively determined — the one permitted rewind |
| `worktrees` key | the path is absent from `git worktree list` |

Artifacts are read from the worktree when one exists, never from the main checkout while a
worktree holds uncommitted work.

### 5. Finish, with cleanup

Ordered; each step gated on the one before it.

1. **Verify the merge.** `gh` when it is installed and the remote is a GitHub host; otherwise
   `git merge-base --is-ancestor`. That fallback must stay reachable on its own — it is the
   only merge evidence on a non-GitHub forge. Not merged → stop, change nothing.
2. **Sync delta specs** into `openspec/specs/`, then move the change into
   `openspec/changes/archive/<date>-<name>/`.
3. **Commit and push the archive** on the base branch in the main checkout. New in KAN-8, and
   the gap that currently leaves this repo sitting on an uncommitted KAN-10 archive move.

   The ticket asks for "commit, merge and push". There is no merge to do in the normal case:
   the change branch was already merged before finish ran — that is step 1's precondition — so
   the archive move is a fresh commit on the base branch itself. The merge step applies only
   when finish is invoked with a non-base branch checked out, in which case it commits there,
   merges into the base branch, and pushes that. Finish never merges the *change* branch; if
   that has not happened, step 1 has already stopped the run.
4. **Clean up the worktrees.** The set is the keys of `worktrees`; when that is absent, scan
   `git worktree list` in each repo for branch `openspec/<name>`. Never guess a path — worktree
   layout differs per repo, and this repo's own live worktree sits in a sibling directory
   (`agents-worktrees/`), not `.worktrees/`.

   Per worktree, check all three first:
   - no uncommitted tracked changes
   - no unpushed commits
   - the local stack is stopped

   Then, and only then: `git worktree remove --force`, `git branch -d`, `git worktree prune`.

   `--force` is for build junk only — `node_modules/`, `build/`, `.next/` — which is why the
   uncommitted-changes check runs first and separately. `git branch -d` is never `-D`: it must
   be allowed to refuse an unmerged branch.

   Already removed is success, not an error. **Any failed check leaves everything alone and
   reports why** — no partial cleanup.
5. **Write FINISHED.**

The stack-stopped check needs a project-specific command, so `.myflow/project.md` gains an
optional `## stop` key naming it (`./gradlew devStop` in Gymie). Absent → the check is skipped,
not failed. This repo has no stack and will not define one.

### 6. Output discipline

Every command ends in the same shape, and prints nothing after it:

```
<1–3 lines: what actually happened>

<absolute paths to anything the human needs to open>

Next:
/myflow-test kan-8-myflow-updates
```

Rules that make that shape true:

- **Absolute paths everywhere** — in handoffs, in generated guides, in IntelliJ commands.
  Never a relative path, never `../sibling`, never a main-checkout path while a worktree holds
  the work.
- **The next command is the last line**, bare and copy-pasteable, with no prose after it. An
  agent cannot drive Claude Code's autocomplete — nothing lets a skill prefill the input box —
  so the last-line convention plus a six-command surface is the whole mechanism.
- **No bodies in chat.** Manual-test guides, diffs and plans are linked, not pasted.
- **The IN_PROGRESS review diff excludes `openspec/`.** The plan was read at STARTED; it should
  not reappear as code to review.

### 7. Repo restructure

One skill per command, named after the command that loads it.

```
skills/
  myflow-start/      ← /myflow-start   (absorbs openspec-propose-fix-superpowers, openspec-propose)
  myflow-do/         ← /myflow-do      (absorbs openspec-apply-fix-superpowers, openspec-apply-change)
  myflow-test/       ← /myflow-test
  myflow-review/     ← /myflow-review
  myflow-finish/     ← /myflow-finish  (absorbs openspec-archive-change, openspec-sync-specs)
  myflow-fast/       ← /myflow-fast
  myflow-info/       myflow-status/    myflow-contracts/
  openspec-explore/  ← /opsx:explore   (kept — no myflow equivalent, touches no state)
```

Deleted: `openspec-full-cycle-superpowers`, `myflow-state-advance` (+ `scripts/state-advance.sh`,
`scripts/test-state-advance.sh`), `openspec-apply-fix-superpowers`,
`openspec-propose-fix-superpowers`, `openspec-propose`, `openspec-apply-change`,
`openspec-archive-change`, `openspec-sync-specs`, `openspec-update-change`.

**The absorptions are not deletions.** `openspec-archive-superpowers` delegates into
`openspec-archive-change` and `openspec-sync-specs`; `openspec-propose-superpowers` delegates
into `openspec-propose`. The delegated content must be inlined into the surviving skill before
its source is removed, or finish and start break.

Commands drop from 18 to 8 in each of `commands/` and `commands-claude/`:
`myflow-start`, `myflow-do`, `myflow-test`, `myflow-review`, `myflow-finish`, `myflow-fast`,
`myflow-info`, `myflow-status`. `commands/opsx-explore.md` survives; the other five `opsx`
command files go.

Flags removed: `automerge` (now a question), `skip-review`, `skip-manual-test`, `skip-propose`,
`propose-only`, `checkpoint`, `commit-during-apply`. `full-panel` survives on `/myflow-do` —
it selects reviewer breadth, not pipeline shape.

### 8. Docs and guards

- `README.md` gains a mermaid transition graph and is rewritten against the five states.
- `myflow-info` and `myflow-status` are rewritten — status reports the five states and the
  next command; info explains the new pipeline from `pipeline.md`.
- `skills/myflow-contracts/pipeline.md` is rewritten: five states, six commands, the
  transition table, git boundaries, the finish contract.
- `CLAUDE.md`, `AGENTS.md`, `skills/README.md` swept for the retired vocabulary.
- `scripts/check-vocabulary.sh` gains the retired twelve-stage values and the retired command
  names as literals, so a half-finished rename fails loudly.
- `scripts/check-references.sh` and `scripts/test-setup.sh` keep working;
  `scripts/test-state-advance.sh` is deleted with the skill it tested.

## Decisions

### What `/myflow-review` does in the five-command flow

**ID:** review-role
**Status:** active
**Chosen:** Keep commit + push + open PR — the `*-manual-review` marker commands are simply
deleted, and reviewing the staged diff becomes the unmarked human gate of IN_PROGRESS. This is
what makes five states map 1:1 onto five commands.
**Considered:** Making `/myflow-review` the manual-review marker and moving commit/push/PR into
`/myflow-finish` — rejected because finish would become one large irreversible step (commit,
push, PR, merge, sync, archive, delete worktrees) with no gate between opening a PR and
destroying the worktree.

### Where a fix leaves the state

**ID:** fix-re-entry
**Status:** active
**Chosen:** `/myflow-do` leaves the state untouched except STARTED → IN_PROGRESS. Removes
`originStage` and the whole fix-re-entry table; the operator re-runs an earlier command if they
want an earlier gate re-opened.
**Considered:** Always dropping back to IN_PROGRESS — safest, since new code could never pass a
gate that ran before it existed, but a one-line fix at REVIEW would cost a full re-walk.
Also considered recording a `staleAfterFix` marker so review could warn — rejected as
reintroducing the bookkeeping field this decision exists to delete.

### How the next command is surfaced

**ID:** next-command-hint
**Status:** active
**Chosen:** Print it bare as the final line of every handoff, with nothing after it. Works in
Cursor and Codex too, and needs no harness feature.
**Considered:** A Claude Code `Stop` hook printing the next command — rejected: it still cannot
prefill the input box, it is Claude-Code-only, and it adds a hook running on every stop in
every project, against this ticket's own simplification goal. Driving autocomplete directly was
investigated and is not possible: nothing lets a running session populate the user's input.

### Fate of the `/opsx:*` commands

**ID:** opsx-removal
**Status:** active
**Chosen:** Delete the five that duplicate pipeline steps; keep `/opsx:explore`, which has no
myflow equivalent and touches no state.
**Considered:** Keeping all of them — they cost nothing at session time, but each is another
file every future rename must sweep. Considered deleting all six — rejected because explore is
a thinking mode, not a pipeline shortcut.

### When `/myflow-review` asks about merging

**ID:** merge-choice-timing
**Status:** active
**Chosen:** Ask before doing any work, then run to completion unattended.
**Considered:** Asking after the PR is open, so the decision is made with test and coverage
results in hand — rejected because it puts a stop in the middle of a run the operator wanted to
leave alone. Considered dropping auto-merge entirely — rejected as losing the one-command path
for trivial changes.

### Skill layout

**ID:** skill-layout
**Status:** active
**Chosen:** One skill per command, named after it — 19 skills → 10.
**Considered:** Minimal churn (rewrite contents, keep filenames) — rejected because the
simplification would be true of the docs but not of the tree, and `/myflow-do` would still load
one of two skills depending on whether it was a fix. Considered a single dispatching skill —
rejected because every command would then load all five stages' instructions, directly against
the token work `2e3f973` just landed.

## Out of scope

- The always-on rule layer. `2e3f973` already reduced it to a 339-word stub; KAN-8 rewrites
  `pipeline.md` beneath it but does not revisit the split.
- Re-running `setup.sh global` on this machine. The installed
  `~/.claude/CLAUDE.md` is stale by two merges — an operator action, not a code change.

## Verification

This repo's own guards, per `.myflow/project.md`:

```bash
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/test-setup.sh
```

Plus a sandboxed installer run, which is the only way to exercise `setup.sh` without writing
to the real home directory:

```bash
SANDBOX="$(mktemp -d)"; HOME="$SANDBOX" ./setup.sh global
```

The vocabulary guard is the regression check for this rename specifically: every retired stage
value and command name must be added to its literal list, and a clean run proves only that
those exact strings are gone — not that the rename is complete. Sweeping each layer by hand is
still what makes it done.
