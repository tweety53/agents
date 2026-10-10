# The gated per-task reviewer — the `fix` path

Loaded during `flow.sdd-tdd` by the load directive under the gated per-task reviewer of
**4. Execute (SDD + TDD)** (`skills/flow/implement.md`), only when a gated reviewer pass closes `fix`.

**The parent applies the fix itself**, never resuming the group's
implementer: one inline round per group carrying every `fix` report of that group's tasks,
recorded as one pair `-role implementer -model <parent model> -effort <parent effort> -agent-id inline` under
`task-<n+n>-implementer-fix-<k>`, the same `+`-joined ids; per task, it stages the
changed paths (`git add -- <the changed paths>` — a pathspec commit reads tracked paths only, so
a fix that adds a file stages first) and commits on the route the branch's push state dictates
(**Panel re-runs**, `skills/flow/review-panel-fix-round.md`) — one new commit on top for a branch the remote
already holds, the guarded, set-aside fold onto `<task-sha>^` for unpushed history only, as that
section states it for a fixup against `<task-sha>`. The
parent re-runs the guard on every sha the fold's rebase rewrote — the on-top route rewrites none, so
its re-run covers nothing — then re-dispatches the reviewer — one
bundle carrying every fixed task of the group, under `task-<n+n>-reviewer-fix-<k>`, the same
convention as the implementer's fix key — each pass on its own range: the on-top route reads its
fix commit's own diff `git diff <fix-commit>^..<fix-commit>`, the fold its rewritten
`git diff <task-sha>^..<new-task-sha>`.
The re-dispatched bundle's Agent-tool `description` is `Tasks <ids>/<N> (re-review)` — `Task <x>/<N> (re-review)` <!-- citations-guard:allow -->
for one task — so the `flow-task-list` mod shows its row as a re-review.

The fix the parent applies is mutation-proved before the reviewer re-dispatches: the parent is
bound by the MUTATION PROOF paragraph (`skills/flow/review-panel.md`) in its exact words — the
same binding **Inline — the parent implements** (`skills/flow/implement.md`) gives the parent on
an `inline` run — and the re-dispatched reviewer's
bundle carries the reported `fix-mutation:` lines beside the fix diff, a mutant nothing kills
failing the gate and sending the task back through this file's fix path.
