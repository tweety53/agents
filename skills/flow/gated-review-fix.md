# The gated per-task reviewer — the stage-end fix

Loaded during `flow.sdd-tdd` by the load directive under the gated per-task reviewer of
**4. Execute (SDD + TDD)** (`skills/flow/implement.md`), only when, at the last group's guard pass,
a gated pass closed `fix` or left a Minor open.

**One fixer for the whole stage**, never the parent and never a resumed implementer: once the last
group's guard has passed and every gated bundle has returned, one dispatch carries every `fix`
report and every Minor a gated pass left open — the gated row of **Every fix is a fresh one-shot**
(`skills/flow/implement.md`), keyed `task-<n+n>-implementer-fix-<k>`, the ids every task it fixes
`+`-joined in plan order, on the decision's `fixer` pair. Its `FIX INPUT:` line names each reviewer
report path, each finding's ref and MUTATION PROOF. Every task commit is pushed by then, so each
fix is one new commit on top per task, carrying that task's `Task-Id:` trailer.

**Its return is one Bash call**: the markers verify, then the guard on each fix commit,
`check-task-commit-fields.sh <canonical> <id> <fix-sha> "" <canonical> <name>`, which
joins every commit carrying the task's `Task-Id:` trailer — a refusal handled as "The record
carries its own corrections" (`skills/flow/implement.md`) states — the push, `flow record status
-change <name> -ref F<n> -status fixed` for each Minor its summary head names `fixed:`, and, when a
task's pass was `fix`, that task's `task-<n>-fix-<k>.diff` of its fix commit's own range,
`git diff <fix-commit>^..<fix-commit>`, the reviewer-dispatch planning commit, the snapshots and
the re-review bundle's `begin`. A `not-fixed:` line is the parent's to judge: withdrawn with its
one clause, or carried to the next fixer.

**One re-review bundle** then carries every task whose pass was `fix`, under
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
