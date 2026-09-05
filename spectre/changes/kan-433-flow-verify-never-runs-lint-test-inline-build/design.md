# flow: verify never runs lint/test inline, build SPA dist once per worktree — design

**Change:** `kan-433-flow-verify-never-runs-lint-test-inline-build`
**Jira:** KAN-433
**Date:** 2026-09-05

## Context

`proposal.md` carries the why; this file carries the how. Bounded change: every flow edited already
exists, and no capability spec under `spectre/specs/` covers these stages, so no spec edit is
planned.

## 1. `## worktree setup` — a new optional project-configuration key

**Contract.** `skills/flow-contracts/project-configuration.md`'s key table gains one row:

> `## worktree setup` — Optional. A fenced command block run once per worktree, from the worktree
> root, immediately after **2. Isolate the workspace** (`skills/flow/implement.md`) creates it and
> before anything else touches the tree — the place for a build that a gitignored, embedded
> artifact needs before the project's first `go test`/`go build` can succeed. Absent means nothing
> runs. Same shape as `## lint` and `## test`: commands in a fenced block, read literally, one
> command per line, in order.

**This repository.** `.flow/project.md` gains a `## worktree setup` section:

```bash verified:stats/Makefile web-build target on main
cd stats && make web-build
```

with the reason beside it: `stats/internal/web/dist/` is gitignored (`.gitignore:23`) and
`stats/internal/web/embed.go`'s `//go:embed all:dist` refuses to compile without it, so a fresh
worktree's first `go test ./...` or `go build ./...` fails until the SPA is built once (kan-389).

## 2. `implement.md` §2 — creation stated inline

"Invoke **superpowers:using-git-worktrees**" is replaced by the commands, in this order:

1. `git worktree add <project>/.worktrees/<name> -b spectre/<name>` from the project's default
   branch HEAD. `<project>/.worktrees/` must be ignored — `git check-ignore -q .worktrees` — and
   where it is not, the run adds it to `.gitignore` and commits that first, on the default branch
   (the one line the skill's Step 1b carried that still matters).
2. `project-get.sh <worktree> "worktree setup"` — exit 0: run every printed command from the
   worktree root, in order, in the foreground; exit 1: nothing declared, say so and continue;
   exit 2: stop the run, relaying the script's own line. A command's non-zero exit ends the
   conductor's turn with `## Question` naming the command and its output — a worktree that cannot
   be set up will fail `flow.verify` anyway, and the operator should see it here.
3. The existing steps, unchanged: the `spectre/changes/<name>/` copy-and-remove, `spectre link`
   where a peer exists, `flow workspace-id`.

A fix run resumes the existing worktree and runs none of the above, as today. The skill's Step 2
(dependency install) and Step 3 (baseline test suite) go: `flow.verify` runs the suite at the end
of the same run and the base-branch guards cover the base.

**Citations to the skill go with it.** `implement.md`'s step table row **2**, `skills/README.md`'s
Basic Workflow map row for step 2, and `skills/flow-contracts/finish-contract-run1.md:331`'s
"per `superpowers:using-git-worktrees`" are re-pointed at **2. Isolate the workspace**
(`skills/flow/implement.md`).

## 3. `verify-and-handoff.md` — the conductor's own boundary

**Verify** gains one paragraph after "This step does not call the project's `create` command":

> **After the panel closes, the conductor edits no source and runs none of the `## lint` or
> `## test` commands itself.** Its work in this stage is `prepare-workspace.sh`, the verifier
> dispatch(es) below, and the ledger render. Any source change from here on makes every slot's
> result stale (**Panel re-runs**, `skills/flow/review-panel.md`), and the only path that changes
> source is a fix run the operator starts.

**The verifier dispatch**, **Recording**: a `## Report` carrying a non-zero exit is re-dispatched
**once** — `-key verify-2` here, `visual-verify-2` in **Visual verification**, the same
`-<worktree basename>` suffix rule — with a prompt that carries the first report verbatim under a
`## Previous attempt` heading. The second report is final: another non-zero exit ends the
conductor's turn with `## Question` naming the failing command and its output, and the operator
resolves it through a fix run. The conductor never runs the failing command itself to check it.
Both dispatches are recorded pairs; the ledger render and stage `end` follow the last one.

**Guardrails** gains: **Never** edit source or run a `## lint`/`## test` command after the panel
closes.

## 4. `review-panel.md` **Panel re-runs**

The stale-result sentence — "a slot's clean result is stale when the rule above required that slot
to re-run and it has not" — is widened: a slot's result is also stale when any commit or working-tree
change to source landed after that slot's last read, from any stage, `flow.verify` included. An
unrecorded post-panel edit is stale by definition, not only one a fix round produced.

## 5. `flow record dispatch`

`stats/cmd/flow/record.go`'s `recordUsage` line `-role is one of: …` lists `verifier`. The
`recordRoles` slice already carries it (`TestRecordAcceptsVerifierRole`); only the usage text lags.
`record_test.go` gains a test asserting that `recordUsage` names every member of `recordRoles`, so
the two cannot drift again.

## Decisions

### Where the SPA build lives

**ID:** worktree-setup-key
**Status:** active
**Chosen:** a new optional `## worktree setup` key in `.flow/project.md`, run by `implement.md` §2
right after `git worktree add` — builds once per worktree, keeps `implement.md` generic, and matches
the issue's "first step of worktree setup" literally.
**Considered:** `## test` opening with `cd stats && make web-build` (the research note's own
wording) — no new key, but builds once per verifier dispatch, not per worktree, and leaves an
implementer's first `go test ./...` failing; `cd stats && make test` in place of the `go test` line —
reuses the Makefile's existing `web-build` prerequisite but hides `-race -count=1` inside the
Makefile and still builds per dispatch.

### What the conductor does on a failed verifier

**ID:** verify-redispatch-once
**Status:** active
**Chosen:** one re-dispatch under `-key verify-2` carrying the first report verbatim, then hand back
with `## Question` on a second non-zero exit — an environmental failure (kan-389's missing `dist`,
a flaky harness) is told apart from a branch defect by the cheapest means, and every attempt is a
recorded pair.
**Considered:** always hand back — simplest, but a kan-389-class failure costs an operator sitting
and a full fix run; dispatch a fix subagent and re-run the affected slots on the delta (the research
note's original cost-lever-2 shape) — correct by construction, but a whole fix round's machinery for
what has so far been one lint slip and one missing build.

## Open questions

None.
