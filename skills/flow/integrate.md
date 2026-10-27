# Integrate (run 1)

Loaded by `skills/flow/SKILL.md` on a **bare** invocation at `IN_PROGRESS` — no argument, or the change's name alone (**Change
name resolution (all `/flow*` commands)**, `skills/flow-contracts/pipeline.md`) — and by **Write
`IN_PROGRESS`** (`skills/flow/verify-and-handoff.md`), which continues here in the same invocation
when the record's `reviewBeforeLanding` is not `true` and no open question is left.

## Deciding which run this is

**`skills/flow-contracts/finish-contract-run1.md` is canonical for every procedure below** — the
base-branch resolution, the preflight checks, the removal sequence and their rationales live in
that file. This file carries only what is specific to *executing* it under `/flow`.

**`integrate-change.sh` runs every mechanical step of run 1 and run 2, in that file's order**
(`skills/flow-contracts/finish-contract-run1.md`), with its stage marks; the parent writes the
commit subjects and the narrative, runs the self-review pass and the Jira calls, and handles each
stop below. **Load `skills/flow-contracts/finish-contract-run1.md` only when** a stop below cites
one of its sections.

**Check guard presence** per **Guard presence check** (`skills/flow-contracts/pipeline.md`),
already run at the top of this invocation.

```bash
integrate-change.sh prepare <main-checkout> <name> --harness <harness> --session-token mf-<literal-token>
```

Every call takes the same `--harness` and `--session-token` its stage marks carry, relays each
guard's own lines, and ends on one machine-readable line — `NEXT:` (exit 0) or `STOP: <kind>` (exit
1); exit 2 is a usage error or an unreadable state record, reported and stopped on — or a guard
missing beside the script, whose steps then run by hand from
`skills/flow-contracts/finish-contract-run1.md` and `skills/flow-contracts/finish-contract-run2.md`. The script's
header (`<agents repo>/scripts/integrate-change.sh`) is canonical for each subcommand's steps and
the exit contract. A stop closes the stage mark it was in `stopped`; the parent re-runs the call
that stopped once the stop is handled.

- **`NEXT:` naming `cleanup`** — every worktree returned `RUN2` → `skills/flow/cleanup.md` <!-- refs-guard:allow -->
- **`STOP: foreign-staged`** → ask the question of **Surface foreign staged work before the
  preflight** (`skills/flow-contracts/finish-contract-run1.md`) once, every repository's lines
  shown together; **Continue** re-runs `prepare` with `--accept-foreign`. A `DRIFT-SWITCHED:` line is relayed as printed and asks nothing: `prepare` already switched that clean checkout back to its default branch and carried on, per **Surface foreign staged work before the preflight** (`skills/flow-contracts/finish-contract-run1.md`).
- **`STOP: migrate`** → stop at `IN_PROGRESS` and report the script's lines, per **Migrate
  retired-layout worktrees before the preflight** (`skills/flow-contracts/finish-contract-run1.md`).
- **`STOP: preflight` or `STOP: base`** → stop, report what the script reported, relay the guard's
  hand-verification procedure per **Hand-verifying a guard verdict** (`skills/flow-contracts/pipeline.md`), and ask
  the operator.
- **`STOP: base-mismatch`** → report the line and ask the operator. A worktree on its default branch whose origin lacks the base another worktree records is not a stop: `prepare` creates that base on its origin from the default branch, records it as the worktree's base with the first and third commands below (the merge base unchanged, since the new base sits at the default branch's commit), leaves moving the branch to the base-moved sync, and prints one `BASE-CREATED:` line.
  The answer that moves the worktree is `git -C <wt> config branch.spectre/<name>.flowBase <b>`, <!-- citations-guard:allow -->
  `git -C <wt> rebase --onto origin/<b> <its recorded merge base>`,
  `flow state add-worktree -C <the change's canonical worktree> <name> <wt> $(git -C <wt> merge-base HEAD origin/<b>)`, then re-run
  `prepare`.
- **`STOP: no-remote`** → per **The routes** (`skills/flow-contracts/finish-contract-run1.md`).

## 1. Check for unfinished work

`prepare` undoes an archive a stopped run 1 left uncommitted, decides once whether the change is
already archived and whether it carries new work, and runs the unfinished-work gate, per **Run 1 —
the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`).

- **`STOP: archive-undo`** → ask, per **Run 1 — the branch is not merged**
  (`skills/flow-contracts/finish-contract-run1.md`).
- **`STOP: unfinished-work`** → **Load `skills/flow/unfinished-work-gate.md`** — the
  breakdown, the three-course prompt and what each course does are there. **Continue** and
  **File or join…** re-run `prepare` with `--accept-outstanding`; **Stop** leaves the mark the
  script closed `stopped`.
  A breakdown naming `no change summary recorded` is closed by recording the summary first — the
  `**Summary:**` bullets the last run printed, or, where this session never saw them, bullets
  written from the change's `tasks.md` and `narrative.md` — with the command the breakdown names,
  then re-running `prepare`.
- **`STOP: live-check`** → the change's `design.md` names a live check and `live-verification.md`
  beside it is missing or fails `check-live-verification.sh`. Run **Live check**
  (`skills/flow/verify-and-handoff.md`) now, through the stand-in the section states wherever a
  step would write this change's integrate state, then re-run `prepare`. `--accept-outstanding`
  never passes it.
- **`STOP: no-verdict`** → a guard of the gate printed no verdict line, or exited non-zero without
  one: stop and ask, reporting its output. `--accept-outstanding` never passes it.
- **`STOP: visual-verify`** → a worktree printed `VISUAL-VERIFY-MISSING`: the UI about to land was
  never shown to a verifier at this HEAD. It is never a Continue, and `--accept-outstanding` never
  passes it. Run **Visual verification** (`skills/flow/verify-and-handoff.md`) from its `begin`
  mark on each such worktree's current HEAD, take every departure through **The loop**
  (`skills/flow/verify-fix-loop.md`) without asking, then re-run `prepare`. The one other pass is
  an operator who hand-verified that HEAD's UI and calls the verdict wrong: record it with
  `flow record verdict false-positive -change <name> -guard check-visual-verify-dispatched -reason
  "<the operator's reason, verbatim>" -C <canonical-worktree>` right after the stop, and re-run `prepare`. The
  command flags the guard's newest verdict, and `prepare` stops at the first `MISSING` worktree
  so that row is the newest; `prepare` passes only on a flag on that worktree's own row whose
  verdict names its current HEAD (`— HEAD <sha>`), so a later commit or a rebase needs a new
  verifier or a new flag. A flag made at `prepare` does not carry to `land`: the reshape and the
  archive move HEAD, so `land` stops again and the operator flags that HEAD anew. The gate runs on
  an archived re-run with no new work too, again on each worktree a base move rebases in
  `prepare`, before every push `commit` and `land` make on every route — `land`'s stop names the
  worktrees already pushed — and again after a re-sync on a rejected push. So the gate covers
  every push the pipeline makes, and the forge merge of an opened PR is of the verified head; run 2
  `cleanup` adds no check, since it could only report after landing.

## 2. Ask how the branch should land

`prepare` then runs `check-base-moved.sh` and, on `MOVED`, the sync and the scoped
re-verification of **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`). A
`REFUSE`, an exit 2 or an empty resolved set is `STOP: base-moved`, handled as `STOP: preflight`
above. **Load `skills/flow/sync-onto-base.md` only when** the script stops on `sync-conflict` or
`sync`, the rebase is resolved in place there, and `prepare` is re-run. `STOP: guard-test` is a
scoped re-verification that failed: report it and stop at `IN_PROGRESS`. `STOP: rebase-lint` is
the project's `## lint` list failing on, or rewriting tracked files of, the rebased tree, handled
the same way.

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
`NEXT:` line names.

```bash
integrate-change.sh commit <main-checkout> <name> <worktree>="<type>(<module>): <what the implementation does>"... --harness <harness> --session-token mf-<literal-token>
```

`<type>`, `<module>` and `<what the implementation does>` are derived from the reshaped diff — the
diffstat `prepare` printed under each `WORKTREE:` line, one subject per worktree the `NEXT:` line
names. A worktree whose reshaped diff holds nothing outside `<project>/spectre/changes/` and no change directory's
`link.md` has an empty implementation commit, so `NEXT:` names no subject for it and `commit` asks none — nor
for the canonical worktree whose only such path is its `link.md`, whose subject **Finish
contract** fixes (`skills/flow-contracts/finish-contract-run1.md`). On
**Continue** at **1** the planning commit also lists the outstanding work: pass it as
`--outstanding <text>`, which `commit` writes into that commit's body as `Outstanding: <text>` under
its own `chore(spectre): plan <name>` subject. An archived change without new work takes no subject: its `NEXT:` names
`commit` bare.

`commit` reshapes, renders the ledger, runs the two-commit chain per worktree in `## Merge order`,
archives the change on its branch, commits the archive and pushes each reshaped branch
`--force-with-lease` — per **Archive on the change branch**
(`skills/flow-contracts/finish-contract-run1.md`). Every stop there — `reshape`, `commit-split`,
`archive`, `commit-archive`, `push`, `pre-reshape` — is reported with the script's lines and leaves the change at
`IN_PROGRESS`; **`/flow` never passes `--force`** to `spectre archive`.

**`STOP: pre-reshape`** → pinning the worktree's HEAD failed, or origin holds a `refs/flow/pre-reshape/<name>` the local pin does not descend from: report both shas and git's output; never overwrite either pin.

**`STOP: narrative`** → an `In-run pipeline fix:` line in the narrative is off the format **Pipeline defects found mid-run** fixes for it (`skills/flow-contracts/pipeline.md`): rewrite that line to that format and re-run `commit`.

**`STOP: narrative`** also stops on a `flow-fix` row `flow self-review findings -change <name>` lists as `fixed` whose sha no `In-run pipeline fix:` line names: append that line, in that format, and re-run `commit`.

**`STOP: narrative`** also stops on a fix or integrate run whose section `narrative.md` lacks (`check-narrative-runs.sh`, the check `flow.write-in-progress` runs): append the missing section and re-run `commit`.

## 4. Archive and run the self-review pass

Per **Run the self-review pass** (`skills/flow-contracts/finish-contract-run1.md`), canonical for it.
Hold the bundle's stdout, then `## Session narrative`, in memory; run the pass over it; then
commit the report in `<canonical-worktree>`, not pushing here:

```bash
land-self-review-report.sh "<canonical-worktree>" "spectre/<name>" \
  "docs(self-review): <name> self-review report" \
  docs/self-review/<name>-self-review.md
```

A branch mismatch or a commit that FAILS is reported and stops the run at `IN_PROGRESS`. A staged
index holding anything beyond the chain's own path refuses the commit the same way
(`LAND-FOREIGN-STAGED`).
A `LAND-UNTRACED-FINDING`, `LAND-DEFERRED-FINDING` or `LAND-REPORT-INVALID` refusal is not a stop: resolve it per step 7 of `/flow-self-review` (`skills/flow-self-review/SKILL.md`) and land again.

A product defect the pass met stops it at the end of step 2 of `/flow-self-review`, closes that mark
`-outcome stopped` instead, and the run stops before step 5 on the handoff block **the sweep**
(`skills/flow-contracts/known-bugs.md`) defines:

```bash
flow stage end -command '/flow' -stage flow.self-review -outcome stopped <name>
```

## 5. Take the chosen route, write the state, and transition Jira

Call it with the parent's own shell outside every worktree in the set, per check 6 of **Worktree
cleanup** (`skills/flow-contracts/finish-contract-run2.md`).

```bash
integrate-change.sh land <main-checkout> <name> [--route "<route>"] --harness <harness> --session-token mf-<literal-token>
```

`--route` carries the operator's answer to the landing question; without it `land` takes the
configured default. `land` closes `flow.self-review`, runs the route of **The routes**
(`skills/flow-contracts/finish-contract-run1.md`) per worktree in `## Merge order` — the
self-review report's sha re-read before each landing push and its `fixed` rows after, per step 5
of `/flow-self-review` (`skills/flow-self-review/SKILL.md`) — and writes `IN_PROGRESS`.

`STOP: landing-route` — no route given or configured — asks the landing question of step 2, then
re-runs `land` with `--route`.

`STOP: loop-landing` — on the merge-and-push route, before each landing push,
`<agents repo>/scripts/check-loop-landing.sh <canonical worktree> spectre/<name>` refused (or could not answer)
the self-review loop run on the change's own branch: nothing more is pushed to the base, the change
stays `IN_PROGRESS`, and the `LOOP-LANDING-REFUSED` lines are relayed — finish the loop they name,
then re-run `land`. A branch with no loop rows answers `LOOP-LANDING-NA` and passes.

**Every git step here can fail, and none of them may fail silently.** `STOP: push`, `STOP: sync`,
`STOP: sync-conflict`, `STOP: guard-test`, `STOP: rebase-lint`, `STOP: self-review-report` and `STOP: state` are
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

```
## Branch integrated — waiting on the merge | merged and waiting on run 2

**Change:** <name>
**Route:** pull request | merged and pushed | manual
**PR:** <prUrl> | none — merged directly | none — you are handling it
**Outstanding:** <what step 1 reported and the operator integrated over> | none
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)

<what the operator must do before the next run>

Next:
/clear
/flow <name>
```

The `**Learning:**` and `**Prompt tuning:**` lines are not printed here — only the Finished handoff (`skills/flow/cleanup.md`) carries them.

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
