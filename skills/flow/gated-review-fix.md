# The gated per-task reviewer — the stage-end fix

Loaded during `flow.sdd-tdd` by the load directive under the gated per-task reviewer of
**4. Execute (SDD + TDD)** (`skills/flow/implement.md`), only when, once the last group's guard has
passed and every gated bundle has returned, a gated pass closed `fix` or left a Minor open.

**One fixer per group of at most twenty findings, in series**, never the parent and never a
resumed implementer: once the last group's guard has passed and every gated bundle has returned,
the `fix` reports and the Minors a gated pass left open are split, in plan order, into groups of
whole tasks each carrying at most twenty findings — a task carrying more is a group alone — and one
fresh dispatch per group carries its tasks' reports and Minors, each launched once the previous
one's return has run, since every group shares one worktree — the gated row of **Every fix is a
fresh one-shot** (`skills/flow/implement.md`), keyed `task-<n+n>-implementer-fix-<k>`, the ids of
the tasks its group fixes `+`-joined in plan order, on the decision's `fixer` pair. Its `FIX INPUT:` line names each reviewer
report path, each finding's ref and MUTATION PROOF. Every task commit is pushed by then, so each
fix is one new commit on top per task, carrying that task's `Task-Id:` trailer.

**Each fixer's return is one Bash call**: the markers verify, then the guard on each fix commit,
`flow-guard check-task-commit-fields <canonical> <id> <fix-sha> "" <canonical> <name>`, which
joins every commit carrying the task's `Task-Id:` trailer — a refusal handled as "The record
carries its own corrections" (`skills/flow/implement.md`) states — the push, `flow record status
-change <name> -ref F<n> -status fixed` for each Minor its summary head names `fixed:`, and, when a
task's pass was `fix`, that task's `task-<n>-fix-<k>.diff` of its fix commit's own range,
`git diff <fix-commit>^..<fix-commit>`, the snapshots and
the re-review bundle's `begin`. A `not-fixed:` line is the parent's to judge: withdrawn with its
one clause, or carried to the next fixer. A clause claiming a test passes or the finding is already
green names the run — its command and exit — else the line is carried to the next fixer, and
`flow record status` refuses a green withdrawal with no exit.

**One re-review bundle**, once the last fixer's return has run, carries every task whose pass was `fix`, under
`task-<n+n>-reviewer-fix-<k>`, the same convention as the fixer's key, each pass on its own fix
diff, its findings file `findings-gated-<n+n>-fix-<k>.jsonl`. The re-dispatched bundle runs on its **Allowed pairs** row (`skills/flow-contracts/model-policy.md`),
its prompt opening with the line `Role: gated-re-reviewer` in place of the original bundle's.
The re-dispatched bundle's Agent-tool `description` is `Tasks <ids>/<N> (re-review)` — `Task <x>/<N> (re-review)` <!-- citations-guard:allow -->
for one task — so the `flow-task-list` mod shows its row as a re-review. It carries the fixer's
`fix-mutation:` lines beside each fix diff, a mutant nothing kills failing the gate and sending
the task back through this file's fix path.

A clean re-review pass closes each finding its task's earlier pass raised, `flow record status
-change <name> -ref F<n> -status fixed`, and ticks its task; a finding a re-review pass raises is
recorded as "The gated per-task reviewer" (`skills/flow/implement.md`) records a first pass's, and
a `fix` sends its task through this file again under the next `<k>`.
