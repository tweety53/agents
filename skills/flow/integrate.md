# Integrate (run 1)

Loaded by `skills/flow/SKILL.md` on a **bare** invocation at `IN_PROGRESS` — no argument, or the change's name alone (**Change
name resolution (all `/flow*` commands)**, `skills/flow-contracts/pipeline.md`) — and by **Write
`IN_PROGRESS`** (`skills/flow/verify-and-handoff.md`), which continues here in the same invocation
when the record's `reviewBeforeLanding` is not `true` and no open question is left.

## Deciding which run this is

**`skills/flow-contracts/finish-contract-run1.md` is canonical for every procedure below** — the
base-branch resolution, the preflight checks, the removal sequence and their rationales live in
that file. This file carries only what is specific to *executing* it under `/flow`.

**`flow-guard integrate-change` runs every mechanical step of run 1 and run 2, in that file's order**
(`skills/flow-contracts/finish-contract-run1.md`), with its stage marks; the parent writes the
commit subjects and the narrative, runs the self-review pass and the Jira calls, and handles each
stop below. **Load `skills/flow-contracts/finish-contract-run1.md` only when** a stop below cites
one of its sections.

**Check guard presence** per **Guard presence check** (`skills/flow-contracts/pipeline.md`),
already run at the top of this invocation.

**A change to `<agents repo>`'s own tooling integrates with the branch's copy of it.** When the
canonical repository is `<agents repo>` and `git -C <canonical-worktree> diff --name-only
<base>...HEAD` names a path under `scripts/`, `stats/` or `skills/` — the three trees whose
installed copies a run executes — every `flow-guard integrate-change` call this file and
`skills/flow/cleanup.md` make resolves to the branch's own copy and runs with a `flow` built from
the branch, both first on `PATH`:

```bash
cd <canonical-worktree>/stats && go build -o <scratch>/bin/flow ./cmd/flow
PATH=<scratch>/bin:<canonical-worktree>/scripts:$PATH flow-guard integrate-change prepare <main-checkout> <name> --harness <harness> --session-token mf-<literal-token>
```

The installed script and CLI predate the change, and the script builds its guards from its own
checkout (`<agents repo>/scripts/flow-guard`) while taking `flow` from `PATH`.

```bash
flow-guard integrate-change prepare <main-checkout> <name> --harness <harness> --session-token mf-<literal-token>
```

Every call takes the run's `--harness` and `--session-token`, relays each
guard's own lines, and ends on one machine-readable line — `NEXT:` (exit 0) or `STOP: <kind>` (exit
1); exit 2 is a usage error or an unreadable state record, reported and stopped on — or a guard
missing beside the script, whose steps then run by hand from
`skills/flow-contracts/finish-contract-run1.md` and `skills/flow-contracts/finish-contract-run2.md`. The script's
header (`<agents repo>/scripts/flow-guard integrate-change`) is canonical for each subcommand's steps and
the exit contract. A stop closes the stage mark it was in `stopped`; the parent re-runs the call
that stopped once the stop is handled. Three stops in a row of one call take **Three failed
rounds** (`skills/flow-contracts/pipeline.md`), applied inline from every earlier `STOP:` line — only
stops the parent can cure itself. A stop that exists to ask the operator — `STOP: main-checkout`,
`STOP: base-mismatch` — is never counted toward it: it is asked as its bullet below says, every time.

- **`NEXT:` naming `cleanup`** — every worktree returned `RUN2` → `skills/flow/cleanup.md` <!-- refs-guard:allow -->
- **`STOP: main-checkout`** → stop at `IN_PROGRESS` and report every repository's lines together,
  per **Surface foreign staged work before the preflight**
  (`skills/flow-contracts/finish-contract-run1.md`); the operator cleans the main checkout and
  re-runs. A `DRIFT-SWITCHED:` line is relayed as printed and asks nothing: `prepare` already switched that clean checkout back to its default branch and carried on, per **Surface foreign staged work before the preflight** (`skills/flow-contracts/finish-contract-run1.md`).
- **`STOP: preflight` or `STOP: base`** → stop, report what the script reported, relay the guard's
  hand-verification procedure per **Hand-verifying a guard verdict** (`skills/flow-contracts/pipeline.md`), and ask
  the operator.
- **`STOP: base-mismatch`** → report the line and ask the operator. A worktree on its default branch whose origin lacks the base another worktree records is not a stop: `prepare` creates that base on its origin from the default branch, records it as the worktree's base with the first and third commands below (the merge base unchanged, since the new base sits at the default branch's commit), leaves moving the branch to the base-moved sync, and prints one `BASE-CREATED:` line.
  The answer that moves the worktree is `git -C <wt> config branch.spectre/<name>.flowBase <b>`,
  `git -C <wt> rebase --onto origin/<b> <its recorded merge base>`,
  `flow state add-worktree -C <the change's canonical worktree> <name> <wt> $(git -C <wt> merge-base HEAD origin/<b>)`, then re-run
  `prepare`.
- **`STOP: no-remote`** → per **The routes** (`skills/flow-contracts/finish-contract-run1.md`).

## 1. Check for unfinished work

`prepare` runs the unfinished-work gate, per **Run 1 —
the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`).

- **`STOP: unfinished-work`** → **Load `skills/flow/unfinished-work-gate.md`** — the run
  finishes the unfinished work itself, asks nothing, and re-runs `prepare`, per its
  **Finish it — never a prompt**.
- **`STOP: no-verdict`** → a guard of the gate printed no verdict line, or exited non-zero without
  one: stop and ask, reporting its output.

## 2. Ask how the branch should land

`prepare` then runs `flow-guard check-base-moved` and, on `MOVED`, the sync of **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`). A
`REFUSE`, an exit 2 or an empty resolved set is `STOP: base-moved`, handled as `STOP: preflight`
above. **Load `skills/flow/sync-onto-base.md` only when** the script stops on `sync-conflict` or
`sync`, the rebase is resolved in place there, and `prepare` is re-run. `STOP: rebase-lint` is
the project's `## lint` list failing on, or rewriting tracked files of, the rebased tree: report it and stop at
`IN_PROGRESS`.

The `ROUTE:` line carries the project's `## default landing route`. **A resolved default skips
the question entirely** — take that route without asking, and say so in the handoff (`Route:
<route> — from this project's configured default, not asked`). Only `ROUTE: ask` falls back to
asking:

> **How should this branch land?**
> - **Open a pull request** *(default, recommended)*
> - **Merge and push**
> - **Handle it manually**

Having asked once, run to completion without asking again.

**Report an existing PR before asking.** If a PR exists for this branch — from `prUrl`, or a PR CLI
when one is usable — say so, open or closed-unmerged, before the operator answers.

## 3. Commit the staged work

**Load `skills/flow-contracts/git-boundaries-commit-chain.md`** before writing the subjects below.

**Append this run's own narrative first**, the same append `flow.write-in-progress` <!-- refs-guard:allow -->
(`skills/flow/verify-and-handoff.md`) makes, heading `## <YYYY-MM-DD> — integrate run`, covering
this run's preflight, the unfinished-work gate and the rebase — into the `narrative.md` the
`NEXT:` line names. An invocation that already appended its section at `flow.write-in-progress`
extends it instead: ` — integrate run` is appended to that section's heading and the preflight,
gate and rebase lines go under it.

```bash
flow-guard integrate-change commit <main-checkout> <name> <worktree>="<type>(<module>): <what the implementation does>"... --harness <harness> --session-token mf-<literal-token>
```

`<type>`, `<module>` and `<what the implementation does>` are derived from the reshaped diff — the
diffstat `prepare` printed under each `WORKTREE:` line, one subject per worktree the `NEXT:` line
names. A worktree whose reshaped diff is empty has an empty implementation commit, so `NEXT:` names no
subject for it and `commit` asks none.

`commit` reshapes, renders the ledger, runs the guarded commit chain per worktree in `## Merge order`
and pushes each reshaped branch `--force-with-lease` — per **Run 1 — the branch is not merged**
(`skills/flow-contracts/finish-contract-run1.md`). Every stop there — `reshape`, `commit-two`,
`push` — is reported with the script's lines and leaves the change at
`IN_PROGRESS`.

## 4. Run the self-review pass

Per **Run the self-review pass** (`skills/flow-contracts/finish-contract-run1.md`), canonical for it.
Write the bundle's stdout, then `## Session narrative`, to `<sr-dir>/bundle.md`, never printing it,
and run the pass over it. No report is written or committed.

A product defect the pass met stops it at the end of step 2 of `/flow-self-review`, closes that mark
`-outcome stopped` instead, and the run continues straight into the fix run **the sweep**
(`skills/flow-contracts/known-bugs.md`) names, in the same invocation, before step 5:

```bash
flow stage end -command '/flow' -stage flow.self-review -outcome stopped <name>
```

## 5. Take the chosen route, write the state, and transition Jira

Call it with the parent's own shell outside every worktree in the set.

```bash
flow-guard integrate-change land <main-checkout> <name> [--route "<route>"] --harness <harness> --session-token mf-<literal-token>
```

`--route` carries the operator's answer to the landing question; without it `land` takes the
configured default. Before `land`, the session ends the
self-review stage itself — `land` no longer ends it:

```bash
flow stage end -command '/flow' -stage flow.self-review -outcome completed <name>
```

`land` checks that row with `flow stage completed -C <main-checkout> -stage flow.self-review <name>`
and stops at `STOP: self-review-report` until it holds; it then runs the route of **The routes**
(`skills/flow-contracts/finish-contract-run1.md`) per worktree in `## Merge order` — the
self-review's `fixed` rows re-read after each landing push, per step 5
of `/flow-self-review` (`skills/flow-self-review/SKILL.md`) — and writes `IN_PROGRESS`.

`STOP: landing-route` — no route given or configured — asks the landing question of step 2, then
re-runs `land` with `--route`.

**Every git step here can fail, and none of them may fail silently.** `STOP: push`, `STOP: sync`,
`STOP: sync-conflict`, `STOP: rebase-lint`, `STOP: self-review-report` and `STOP: state` are
reported with the command's own output, leaving the change at `IN_PROGRESS`; `STOP: base` and
`STOP: no-remote` are handled as in **Deciding which run this is** above; a conflict on the merge-and-push route's one re-sync is resolved per
**Sync the branch onto the base** (`skills/flow/sync-onto-base.md`) and `land` re-run as the one
retry.

**`STOP: pr`** — no usable `gh`: print the forge's create-PR URL and ask whether it was opened.
**Human confirmation is a legitimate substitute for an API probe** on a forge with no usable CLI. If
the answer is No, leave `prUrl` null and say what to do next. Either answer re-runs `land` adding
`--pr-url <worktree>=<url>` or `--pr-url <worktree>=none` for the worktree the stop names, every
earlier `--pr-url` kept.

**`JIRA: transition <KEY> to In Review`** → **Load `skills/flow-contracts/jira-integration.md`** and
**`skills/flow-contracts/jira-integration-finish.md`** — its **In Review is not tied to a pull
request** governs the transition — and transition the issue per **Transitions**
(`skills/flow-contracts/jira-integration.md`): after the state write, never before, never
blocking.

## Handoff

````
## Branch integrated — waiting on the merge | merged and waiting on run 2

**Change:** <name>
**Route:** pull request | merged and pushed | manual
**PR:** <prUrl> | none — merged directly | none — you are handling it
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)

<what the operator must do before the next run>

Next:
/clear
```
/flow <name>
```
````

| Route taken in step 5 | Heading |
|--------------------|---------|
| pull request | *waiting on the merge* |
| **merge and push** | *merged and waiting on run 2* — continue below |
| manual | *waiting on the merge* |

Where the route is not certain — a run resumed after a partial failure — take the answer from the
merge-status test in **The block each state renders**
(`skills/flow-contracts/handoff-blocks.md`) rather than assuming.

## After merge-and-push specifically

`land` continues, within the same invocation and without a further command from the operator,
into run 2; handle its output by `skills/flow/cleanup.md` exactly as written. Nothing external
blocks this route.

## After open PR or manual specifically

Stop after the route completes, printing the handoff above. Each of these two routes needs an
action outside this command's control before the branch merges. The next bare `/flow <name>`
call, once the branch is integrated, runs run 2's cleanup (`skills/flow/cleanup.md`).
