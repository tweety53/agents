# The gated per-task reviewer — the `fix` path

Loaded during `flow.sdd-tdd` by the load directive under the gated per-task reviewer of
**4. Execute (SDD + TDD)** (`skills/flow/implement.md`), only when a gated reviewer pass closes `fix`.

**The parent applies the fix itself**, never resuming the group's
implementer: one inline round per group carrying every `fix` report of that group's tasks,
recorded as one pair `-model <parent model> -effort <parent effort> -agent-id inline` under
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
