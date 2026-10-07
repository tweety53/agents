# Execute — `sdd` dispatch

Loaded during `flow.sdd-tdd` by the load directive of **4. Execute (SDD + TDD)**
(`skills/flow/implement.md`), only when the recorded decision's `execution` is `sdd`. `<shape>`,
`<changeRoot>`, `<principles-path>` and `<canonical-worktree>` are the values that section defines
(`skills/flow/implement.md`). An `inline` run never loads this file (**Inline — the parent
implements**, `skills/flow/implement.md`).

**Dispatch one implementer per group, not per bundle.** The unit is the recorded decision's
`groups` entry — `{bundles, model, effort}`, `bundles` an array of bundle ids from the same
`plan-dispatch-bundles.sh` output above, `model`/`effort` what this group's implementer runs on; a
`null` `groups` field, which only inline execution ever records, never reaches this section, since
inline runs bundles in plan order with no implementer dispatch at all. A group's implementer works
its bundles in plan order, one commit per task, carrying that task's own `Task-Id:` trailer, and a `Build: red` task is bundled with, and
commits with, the partner its `**Squash-with:**` field names.

**Waves — concurrent dispatch of ready groups.** A group is ready when every id in the union of
its bundles' `after <k>:` lines **that is not itself a task of one of the group's own bundles** has
landed — committed and guard-passed, by direct commit or pick.
**At most two implementer dispatches are in flight per wave**, on every `sdd` decision.
A group alone in its wave, with no other group ready alongside it, dispatches into the canonical
worktree and commits directly; two ready groups launch together in one message,
each into its own throwaway worktree created by
`throwaway-worktree.sh create <worktree> <worktree>-wave-group-<g>`, each copy then running the
project's resolved `## worktree setup` command once before its implementer dispatches. A third or
later ready group queues in plan order and launches, into its own throwaway worktree by the same
call, as soon as one of the two in-flight groups is picked — the cap bounds dispatches in
flight, never how many groups may be ready at once. Exit 2 means a step failed, named on stderr, and
nothing is dispatched into that copy.

**As wave members return**, each is cherry-picked onto the change branch in plan order — a member
is picked once every plan-earlier member of its wave is picked. The same
`close-task.sh` call the task-close step (`skills/flow/implement.md`) runs covers each picked
commit, its `-end-commit` recording the picked sha. A pick conflict or a
guard failure is the parent's own to fix — it resolves the conflict or re-commits in the canonical
worktree itself and re-runs the guard — while sibling members, queued groups and
already-ready later waves are unaffected. A copy is removed once its group is picked —
`throwaway-worktree.sh remove <worktree> <worktree>-wave-group-<g>`. A member reporting BLOCKED follows the existing BLOCKED handback. The
one-implementer-per-worktree rule is untouched: each wave member has its own worktree.

**Gather one context bundle per group, immediately before that group's implementer goes out.**
Take `<g>` and the union of the ids from the `bundle <k>: <ids>` lines `plan-dispatch-bundles.sh`
printed for every bundle in the group, comma-separated:

```bash
mkdir -p <worktree>/.superpowers/sdd
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  <worktree>/.superpowers/sdd/dispatch-context-group-<g>.md <id>[,<id>…] <canonical-worktree> <shape>
```

A non-zero exit — including the guard being absent — is reported, and
dispatching proceeds without a context bundle: the dispatch prompt carries the change's proposal,
design, engineering principles and the group's own tasks inline instead; an implementer's context
bundle never gates a run (the panel's bundle failure is its own prompt, in `skills/flow/review-panel.md`). Confirm the bundle was actually written (`test -f
<worktree>/.superpowers/sdd/dispatch-context-group-<g>.md`) and report plainly if it is not.
**Never read the bundle back into this context** — `test -f` is the whole check; its content is the
implementer's input, not the dispatcher's. Report the script's stderr line for this stage (`bundle
unchanged — reusing …` or `bundle rebuilt — …`) as part of this stage's own reporting.

The sixth argument scopes the group's `## tasks.md` section to the plan header and the named
tasks' blocks; a named id the plan does not carry is
exit 2, a plan defect reported like a missing `**Files:**` field. The panel's and the fix
subagent's bundles (`skills/flow/review-panel.md`) pass an empty sixth argument and keep the whole plan.
