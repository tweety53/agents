# Reachability check and withdrawal

Loaded by `skills/flow/brainstorm-planner.md` when the linked issue carries `flow-fix` or
`flow-cost`, and by `skills/flow/resume.md` on a planless `STARTED` resume.

### The reachability check

**A run on a filed fix/cost finding verifies the defect still exists before planning.** When the
linked issue's labels carry `flow-fix` or `flow-cost`, the checklist opens, before any design
question, with a reachability check against the resolved base (the base the `flow.kickoff`
worktree was created from): state the finding's defect as a claim the tree can answer, then run
the cheapest thing that answers it — the guard the finding names, the contract section it says is
missing, the behaviour it reports. A finding the base already delivers — the guard passes, the
line is already there — ends the run: report the evidence, the command run or the line quoted,
and stop before convergence; nothing is planned, and the issue is the operator's to close. The
end then offers the withdrawal route (**The withdrawal route**, below) —
the evidence already showed the change has nothing to plan. A
finding that still reproduces plans as normal, its evidence carried into proposal.md's `## Why`
when **C** writes it.

### The withdrawal route

A change abandoned before planning — nothing implemented, nothing merged — ends `FINISHED` with
its record's `withdrawn` field set (`skills/flow-contracts/state-file.md`), its worktree deleted
and its branches gone, instead of littering every later candidate resolution and `/flow-status`
report with a `STARTED` record nothing can retire. The route runs inside the creating run —
inside the already-open `flow.brainstorm` stage, which closes with the withdrawal as its work —
and no command exists for it.

**Two offer points, both explicit asks:**

- **The reachability check's end** — after the evidence report, the run asks **Withdraw this
  change?** — **Yes — withdraw** *(recommended)* / **No — leave it at `STARTED`**. Yes is
  recommended because the run just proved nothing should be planned.
- **A resume of a planless `STARTED` change** — the `total == 0` case of **Resuming at `STARTED`** (`skills/flow/resume.md`) — opens with
  **Resume brainstorming, or withdraw?** — **Resume brainstorming** *(default, recommended)* /
  **Withdraw this change**. Resume is the default because the operator invoked `/flow <name>` to
  continue.

The ask names exactly what will be deleted — the worktree's absolute path, the local branch
`spectre/<name>`, the remote branch `origin/spectre/<name>` — and that the record ends
`FINISHED, withdrawn`. The explicit answer is the only consent any of those deletions get.

Both offers stay asked under the `## decisions: recommended` mode (**Auto-resolution**,
`skills/flow-contracts/operator-prompts.md`) — the explicit answer is consent to an irreversible
deletion, and the mode never covers that.

**Steps, in order — git first, record last**, so a crash leaves a re-runnable route rather than a
terminal record over a live worktree. Every step tolerates the previous run's landed work — a
worktree already gone, a branch already deleted, a remote delete already done — so the re-run
reaches the same end from any crash point:

1. Per worktree of the resolved set (**Resolving a change's worktrees**,
   `skills/flow-contracts/worktree-resolution.md`): `git status --porcelain` must be empty —
   anything else stops the route with the output shown, before any deletion. A worktree already
   gone — a previous run's remove landed and a later step failed — is skipped, not a stop: the
   route re-runs to the same end.
2. `git -C <main-checkout> worktree remove --force <path> && git -C <main-checkout> worktree
   prune` — both from the worktree's own main checkout, never `-C <worktree>`: the remove deletes
   that directory, and a prune pointed inside it fails `fatal: cannot change to '<path>'` on
   every run. A worktree the previous run already removed reports `fatal: ... is not a working
   tree`: that is the previous run's landed work — run the prune regardless and continue.
3. `git branch -D spectre/<name>` in the repository — an already-deleted branch reports
   `error: branch not found`: the previous run's deletion landed, and the route continues, exactly
   as step 4 tolerates a failed remote delete.
4. `git push origin --delete spectre/<name>`. A failure here is one reported line and the route
   continues — a remote branch that would not delete never blocks the record.
5. One state write, read-merge-write like every write (**State file**,
   `skills/flow-contracts/state-file.md`): `state` `FINISHED`, `withdrawn` `true`, `worktrees`
   `{}`. The write needs a daemon that knows the `withdrawn` field: an older daemon refuses the
   payload as an unknown field, the write falls back to the journal, the entry is retired as
   definitively invalid, one ⚠ line names it — the record unchanged, the route re-run once the
   daemon is current.

The handoff is terminal and names no next command (**Handoff output**,
`skills/flow-contracts/pipeline.md`).

**Scope.** `STARTED` only: an `IN_PROGRESS` change lands through integrate even when the work
disappoints. No Jira action — closing the linked issue, if any, is the operator's own act. The
route marks nothing of its own: the stage it runs inside carries the marks, as any other stage's
work does.
