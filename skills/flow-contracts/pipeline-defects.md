# Pipeline defects found mid-run

Loaded by the load directive of **Pipeline defects found mid-run** (`skills/flow-contracts/pipeline.md`),
only when a run hits a defect in the pipeline itself. Every section name below without a path is a
section of `skills/flow-contracts/pipeline.md`.

**A `/flow` run that hits a defect in the pipeline itself fixes it within the
run, unless the fix is `big`.** The pipeline is everything `<agents repo>` ships — a skill, a
contract, a guard, the `flow` CLI or flowd, a rule — and any saved memory that encodes its
behaviour. For this section alone `<agents repo>` is a repository the run was given
(**Reporting back**, `rules/agent-baseline.md`), whichever project the run is in.

**Predict the class before any work, from the blast radius — never from a plan.** Count the
files the fix must touch: the defective file, its test, and every file `grep -rl` finds in
`<agents repo>` citing the name or section the fix changes. The fix is `big` when that count is 60 or more. A `big` fix is
deferred to self-review, and so is one found `big` once under way, which stops there. **Nothing
else is ever deferred** — not "after the loop", not to the self-review, not to the operator: a
non-`big` fix lands before the run's final report, and `flow-guard integrate-change commit` stops on a
`deferred` line under 60 files.

**A fix that reaches outside `<agents repo>` is never `big` for that alone** — a fix to the
project's own dev tooling (`<project>/.flow/project.md`, build scripts, dev-stack or test-harness config)
is made like any other. It is committed on the change's own branch in that repository's worktree,
where it rides the run's route; a repository the change holds no worktree in takes a worktree off
its default branch, landed by its own `## default landing route`. Production stays out of reach
(`rules/never-touch-production.mdc`).

**A design choice never makes a fix `big`** — the fixer weighs the options and
takes the recommended one without asking, whether the choice is where a check lives, a name, a
scope, a mechanism or whether a guard or a test is added: the fix takes the simplest option
and notes it, the smallest change that closes the defect, no refactor and no detour.

`<agents-base>` is `<agents repo>`'s own default branch —
`git -C <agents repo> symbolic-ref --short refs/remotes/origin/HEAD` with its `origin/` prefix
dropped — never the project's `<default-branch>`, which `<agents repo>` may not carry.

**Every dispatch in the loop is one-shot.** No `SendMessage` to a finished child, and no child
parks a question for a resume. **Each in-run fix runs the fix of 3. Fix every `flow-fix` finding that is not
`big`** (`skills/flow-self-review/SKILL.md`), canonical for its fixer pair,
the paragraphs every prompt carries, the no-progress escalation, and each item the steps below
cite. One loop carries one defect, `<k>` one more than the highest `<n>` of any
`pipeline-fix-*-<n>` change record in flowd, worktree name, ref, or `flow.loop-k-used` mark —
whatever change it was for — so `<k>` is unique across
`<agents repo>` and no two loops, concurrent or later, share a key (`flow-guard pipeline-fix start`
refuses any other `<k>`); its name
`pipeline-fix-<name>-<k>`, `<name>` the running change, stands for `self-review-<name>` there —
the change, its branch, its worktree and `<fix-branch>` — and `origin/<agents-base>` for
`<fix-base>`; the parent is that file's "this session".

1. **Fix** — the parent creates the fix's own worktree with
   `<agents repo>/scripts/flow-guard pipeline-fix start <name> <k>` — never the main checkout, never the
   change's own worktree. `pipeline-fix start`
   comes before any edit, a parent's hand fix included: `pipeline-fix land` refuses a worktree not on
   the loop branch `start` creates. It dispatches one fixer, `pipeline-fix-<k>-fix`, its prompt carrying the defect, its evidence and the counted blast
   radius, as item 1 there states. The prompt carries its `Key: pipeline-fix-<k>-fix` as a line
   of its own, from which `flow-guard hook enforce-loop-dispatch` records the row under the loop's change;
   the parent never runs `flow record dispatch begin` or `end` for it under its own change. The fixer also adds one test or guard that fails without the fix,
   and updates a saved memory that encoded the defect.
2. **Review once, primary only** — as item 2 there states it, on the `loop-review-1` and `loop-review-r` rows, keys `pipeline-fix-<k>-review-1`, `pipeline-fix-<k>-fix-r<r>` and `pipeline-fix-<k>-review-r<r>`.
3. **Land** — it lands by
   `<agents repo>`'s `## default landing route`
   (`flow-guard project-get <agents repo> 'default landing route'`), without asking. Merge and push is
   `<agents repo>/scripts/flow-guard pipeline-fix land <worktree> <important|minor>`, the last argument the defect's grade as **Record each one** below defines it, run with a Bash `timeout: 600000` —
   a landing that moved `<agents repo>/stats/` restarts flowd inside the call, and a kill partway leaves it down.
   Its `red:` lines feed
   `pipeline-fix-<k>-fix-v<v>` as item 3 there states; three red lands take **Three failed rounds** (`skills/flow-contracts/pipeline.md`).

The loop runs in the background while the run's own work continues; the run still never ends a
turn with one of its children in flight.

**Record each one.** Every in-run fix, landed or deferred, is one line in the run's narrative —
`narrative.md` on `/flow` (**Write `IN_PROGRESS`**, `skills/flow/verify-and-handoff.md`) —
and one decision bullet in the run's summary
(**Summary and live-stack line, before every handoff**, `skills/flow-contracts/pipeline.md`):

```text
In-run pipeline fix: <agents sha | deferred> — <the defect, one line> (blast radius <N> files)
```

The sha, there and in the store row below, is the defect's task commit — its first, when the plan
holds several — re-read by its subject on `<agents-base>` after landing, as item 2 of that fix
states. A landed fix is also one row in the flow store, recorded once its sha is on
`<agents-base>`:

```bash
flow self-review finding -change <name> -angle flow-fix -disposition fixed -ref <sha> -blast-radius <N> -severity <important|minor> -note '<the defect, one line>'
```

`-severity` grades the defect: `important` when it blocked a step of the run, `minor` otherwise.

On the merge-and-push route `flow-guard pipeline-fix land` records that row itself once its push succeeds — `<name>` read off the branch `pipeline-fix-<name>-<k>`, `-ref` the loop's first commit, `-blast-radius` the number of files the loop's diff touches, `-note` that commit's subject — in the flowd project of the directory it runs from, so the parent runs it from the running change's own working directory; on any other route the parent records the row with the command above.

A deferred fix records no row here; the self-review pass records it once it decides it.
