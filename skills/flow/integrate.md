# Integrate (run 1)

Loaded by `skills/flow/SKILL.md` on a **bare** invocation at `IN_PROGRESS` — no argument, or the change's name alone (**Change
name resolution (all `/flow*` commands)**, `skills/flow-contracts/pipeline.md`) — and by **Write
`IN_PROGRESS`** (`skills/flow/verify-and-handoff.md`), which continues here in the same invocation
when no open question is left.

## Deciding which run this is

**`skills/flow-contracts/finish-contract-run1.md` is canonical for every procedure below** — the
base-branch resolution, the preflight checks, the removal sequence and their rationales live in
that file. This file carries only what is specific to *executing* it under `/flow`.

**`flow-guard integrate-change` runs every mechanical step of run 1 and run 2, in that file's order**
(`skills/flow-contracts/finish-contract-run1.md`), with its `flow.land` mark (run 1) and `flow.cleanup` mark (run 2); the parent writes the
commit subjects, runs the self-review pass and the Jira calls, and handles each
stop below. **Load `skills/flow-contracts/finish-contract-run1.md` only when** a stop below cites
one of its sections.

**Check guard presence** per **Guard presence check** (`skills/flow-contracts/pipeline.md`),
already run at the top of this invocation.

**Load `skills/flow/integrate-own-tooling.md` only when** the canonical repository is `<agents repo>`
and `git -C <canonical-worktree> diff --name-only <base>...HEAD` names a path under `scripts/`,
`stats/` or `skills/` — it carries **A change to `<agents repo>`'s own tooling integrates with the
branch's copy of it**.

```bash
flow-guard integrate-change prepare <main-checkout> <name>
```

Every call takes the run's `--harness` and `--session-token` when given (both inferred otherwise), relays each
guard's own lines, and ends on one machine-readable line — `NEXT:` (exit 0) or `STOP: <kind>` (exit
1); exit 2 is a usage error or an unreadable state record, reported and stopped on — or a guard
missing beside the script, whose steps then run by hand from
`skills/flow-contracts/finish-contract-run1.md` and `skills/flow-contracts/finish-contract-run2.md`. The script's
header (`flow-guard integrate-change`) is canonical for each subcommand's steps and
the exit contract. A stop closes the stage mark it was in `stopped`; the parent re-runs the call
that stopped once the stop is handled. Three stops in a row of one call take **Three failed
rounds** (`skills/flow-contracts/pipeline.md`), applied inline from every earlier `STOP:` line — only
stops the parent can cure itself. A stop that exists to ask the operator — `STOP: main-checkout`,
`STOP: base-mismatch` — is never counted toward it: it is asked as its bullet in **Stops** below says, every time.

- **`NEXT:` naming `cleanup`** — every worktree returned `RUN2` → `skills/flow/cleanup.md` <!-- refs-guard:allow -->
- **Any other `STOP:` line** → handled per **Stops** below.

## 1. Check for unfinished work

`prepare` runs the unfinished-work gate, per **Run 1 —
the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`).

- **`STOP: unfinished-work`** → **Load `skills/flow/unfinished-work-gate.md`** — the run
  finishes the unfinished work itself, asks nothing, and re-runs `prepare`, per its
  **Finish it — never a prompt**.

## 2. Ask how the branch should land

`prepare` then reads the base-moved line the preflight printed after `RUN1` and, on `MOVED`, runs the sync of **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`). A
`REFUSE`, an exit 2 or an empty resolved set is `STOP: base-moved`, handled as `STOP: preflight`
(**Stops** below). **Load `skills/flow/sync-onto-base.md` only when** the script stops on `sync-conflict` or
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

```bash
flow-guard integrate-change commit <main-checkout> <name> <worktree>="<type>(<module>): <what the implementation does>"...
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

## 4. Take the chosen route and write the state

Call it with the parent's own shell outside every worktree in the set.

```bash
flow-guard integrate-change land <main-checkout> <name> [--route "<route>"]
```

`--route` carries the operator's answer to the landing question; without it `land` takes the
configured default. `land` runs the route of **The routes**
(`skills/flow-contracts/finish-contract-run1.md`) per worktree in `## Merge order` and writes
`IN_PROGRESS`. It neither needs a completed self-review nor calls `cleanup`: step 5 runs after it.

`STOP: landing-route` — no route given or configured — asks the landing question of step 2, then
re-runs `land` with `--route`.

**Every git step here can fail, and none of them may fail silently** — each `STOP:` line is
handled per **Stops** below.

## 5. Run the self-review pass

Per **Run the self-review pass** (`skills/flow-contracts/finish-contract-run1.md`), canonical for it.
Write the bundle's stdout to `<sr-dir>/bundle.md`, never printing it,
and run the pass over it. No report is written or committed. The pass runs after `land`, on the
branch it landed; its fixes go on `self-review-<name>` and land by `flow-guard pipeline-fix land`
(`skills/flow-self-review/SKILL.md`), never back onto the change's branch.

A product defect the pass met stops it at the end of step 2 of `/flow-self-review`, closes that mark
`-outcome stopped` instead, and the run continues straight into the fix run **the sweep**
(`skills/flow-contracts/known-bugs.md`) names, in the same invocation:

```bash
flow stage end -command '/flow' -stage flow.self-review -outcome stopped <name>
```

Otherwise the session ends the stage itself, before `cleanup` or the handoff:

```bash
flow stage end -command '/flow' -stage flow.self-review -outcome completed <name>
```

When the stage was ended `stopped` first, `land`'s begin no longer holds it open: begin it again before the
pass whose `completed` end follows, with the run's own harness and token:

```bash
flow stage begin -C <main-checkout> -command '/flow' -stage flow.self-review <name>
```

On merge and push, `cleanup` (`skills/flow/cleanup.md`) then refuses with `STOP: self-review-report`
until that row holds.

## Stops

### `prepare`

- **`STOP: main-checkout`** → stop at `IN_PROGRESS` and report every repository's lines together,
  per **Surface main-checkout drift before the preflight**
  (`skills/flow-contracts/finish-contract-run1.md`); the operator cleans the main checkout and
  re-runs. A `DRIFT-SWITCHED:` line is relayed as printed and asks nothing: `prepare` already switched that clean checkout back to its default branch and carried on, per **Surface main-checkout drift before the preflight** (`skills/flow-contracts/finish-contract-run1.md`).
- **`STOP: preflight` or `STOP: base`** → stop, report what the script reported, relay the guard's
  hand-verification procedure per **Hand-verifying a guard verdict** (`skills/flow-contracts/pipeline.md`), and ask
  the operator.
- **`STOP: base-mismatch`** → report the line and ask the operator. A worktree on its default branch whose origin lacks the base another worktree records is not a stop: `prepare` creates that base on its origin from the default branch, records it as the worktree's base with the first and third commands below (the merge base unchanged, since the new base sits at the default branch's commit), leaves moving the branch to the base-moved sync, and prints one `BASE-CREATED:` line.
  The answer that moves the worktree is `git -C <wt> config branch.spectre/<name>.flowBase <b>`,
  `git -C <wt> rebase --onto origin/<b> <its recorded merge base>`,
  `flow state add-worktree -C <the change's canonical worktree> <name> <wt> $(git -C <wt> merge-base HEAD origin/<b>)`, then re-run
  `prepare`.
- **`STOP: no-remote`** → per **The routes** (`skills/flow-contracts/finish-contract-run1.md`).
- **`STOP: no-verdict`** → a guard of the gate printed no verdict line, or exited non-zero without
  one: stop and ask, reporting its output.

### `land`

**Every git step here can fail, and none of them may fail silently.** `STOP: push`, `STOP: sync`,
`STOP: sync-conflict`, `STOP: rebase-lint` and `STOP: state` are
reported with the command's own output, leaving the change at `IN_PROGRESS`; `STOP: base` and
`STOP: no-remote` are handled as in `prepare` above; a conflict on the merge-and-push route's one re-sync is resolved per
**Sync the branch onto the base** (`skills/flow/sync-onto-base.md`) and `land` re-run as the one
retry.

**`STOP: pr`** — no usable `gh`: print the forge's create-PR URL and ask whether it was opened.
**Human confirmation is a legitimate substitute for an API probe** on a forge with no usable CLI. If
the answer is No, leave `prUrl` null and say what to do next. Either answer re-runs `land` adding
`--pr-url <worktree>=<url>` or `--pr-url <worktree>=none` for the worktree the stop names, every
earlier `--pr-url` kept.

## Handoff

````
## Branch integrated — waiting on the merge | merged and waiting on run 2

**Change:** <name>
**Route:** pull request | merged and pushed | manual
**PR:** <prUrl> | none — merged directly | none — you are handling it

<what the operator must do before the next run>

Next:
/clear
```
/flow <name>
```
````

| Route taken in step 4 | Heading |
|--------------------|---------|
| pull request | *waiting on the merge* |
| **merge and push** | *merged and waiting on run 2* — continue below |
| manual | *waiting on the merge* |

Where the route is not certain — a run resumed after a partial failure — test whether the branch
has reached the base branch rather than assuming: proven merged is *merged and waiting on run 2*;
not merged, or inconclusive, is *waiting on the merge*.

## After merge-and-push specifically

After step 5, the same invocation runs `flow-guard integrate-change cleanup` without a further
command from the operator; handle its output by `skills/flow/cleanup.md` exactly as written.
Nothing external blocks this route.

## After open PR or manual specifically

Run step 5, then stop, printing the handoff above. Each of these two routes needs an
action outside this command's control before the branch merges. The next bare `/flow <name>`
call, once the branch is integrated, runs run 2's cleanup (`skills/flow/cleanup.md`).

## Signal → Action

An index of the stops above; each row's cited section is canonical.

| Signal | Action |
|---|---|
| `NEXT:` naming `cleanup` | **Deciding which run this is** |
| `STOP: unfinished-work` | **1. Check for unfinished work** |
| `STOP: base-moved`, `STOP: sync-conflict` | **2. Ask how the branch should land** |
| `STOP: main-checkout`, `STOP: preflight`, `STOP: base`, `STOP: base-mismatch`, `STOP: no-remote`, `STOP: no-verdict` | **Stops** — `prepare` |
| `STOP: push`, `STOP: sync`, `STOP: rebase-lint`, `STOP: state`, `STOP: pr` | **Stops** — `land` |
| exit 2 | **Deciding which run this is** |
