# Execute — `sdd` dispatch

Loaded during `flow.sdd-tdd` by the load directive of **4. Execute (SDD + TDD)**
(`skills/flow/implement.md`). `<changeRoot>` and `<canonical-worktree>` are the values that file
defines (`skills/flow/implement.md`).

**Dispatch one implementer per group.** The unit is the recorded decision's
`groups` entry — its `tasks` (`skills/flow/decider-prompt.md` is canonical for how they are
grouped), its implementer on the `implementer` role's pair. A group's implementer works
its tasks in plan order, one commit per task, carrying that task's own `Task-Id:` trailer.

**Waves — concurrent dispatch of ready groups.** `flow-guard ready-groups <changeRoot>/tasks.md <canonical-worktree> <name>` prints the groups
to launch now, one `launch <g>: <ids> <repository> direct|throwaway` line each, and the
parent launches exactly those — at the stage's first dispatch and at every group boundary
(`skills/flow/implement.md`), never the next group in plan order. A group is ready when every
task its tasks build on **that is not itself a task of the group** has landed — committed and
guard-passed, by direct commit or pick (the guard's header is canonical).
**At most three implementer dispatches are in flight**, and a
repository launches its next groups only once none of its groups is in flight. A `direct`
group — the only one launching in its repository — dispatches into that repository's
canonical worktree and commits directly; `throwaway` groups — two or three of one repository
launching together — each launch into their own throwaway worktree created by
`flow-guard throwaway-worktree create <worktree> <worktree>-wave-group-<g>`, each copy then
running the project's resolved `## worktree setup` command once before its implementer
dispatches. Every printed group launches in one message, whatever its repository. Exit 1 —
no group can launch while one is unlanded — is a plan defect, reported with the lines it
printed; exit 2 stops the run. `flow-guard throwaway-worktree create` exit 2 means a step failed, named on stderr, and
nothing is dispatched into that copy. A group never spans two repositories (`skills/flow/decider-prompt.md`), so each group's worktree is its repository's.

**As wave members return**, each is cherry-picked onto the change branch in plan order — a member
is picked once every plan-earlier member of its repository's wave is picked. The same
`flow-guard close-task` call the task-close step (`skills/flow/implement.md`) runs covers each picked
commit. A pick conflict or a
guard failure goes to a fresh fixer (**Every fix is a fresh one-shot**, `skills/flow/implement.md`)
— a conflict aborted first, `git -C <canonical-worktree> cherry-pick --abort`, then the pick row
dispatched into the canonical worktree, its `FIX INPUT:` naming the member's commits; a guard
failure the re-commit row — and the parent re-runs the guard on what it commits, while sibling
members, queued groups and already-ready later waves are unaffected. A copy is removed once its group is picked —
`flow-guard throwaway-worktree remove <worktree> <worktree>-wave-group-<g> --fold-back`, which first copies the group's implementer reports back into `<worktree>` for its gated review — its own Bash call with literal absolute paths, never chained into a compound command, which the harness's rm safety check refuses. A member reporting BLOCKED follows the existing BLOCKED handback. The
one-implementer-per-worktree rule is untouched: each wave member has its own worktree.
