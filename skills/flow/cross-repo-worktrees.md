# Isolate the workspace — additional worktrees

Loaded during `flow.isolate-workspace` by the load directive of **2. Isolate the workspace**
(`skills/flow/implement.md`), only when the change links a peer, an `## apps` entry names a
repository holding no worktree for this change, `## visual verification` names a
`regression checkout`, or the resolved worktree set spans more than one repository.

**Every worktree this stage creates for a linked peer runs its
repository's `## worktree setup` before anything else touches it**: before its `spectre link`, its
link or planning commit, and any task. Run `project-get.sh <that worktree> "worktree setup"` and
handle it exactly as step 3 of **A. Resolve the change and write `STARTED`**
(`skills/flow/brainstorm.md`) does for the kickoff worktree — only the fenced command lines, from
that worktree's root, in order, in the foreground; exit 1 says the key is absent and continues;
exit 2 stops the run; a command's non-zero exit ends the turn naming the command and its output. A
worktree this stage resumes rather than creates runs nothing here. An `## apps` entry's worktree
below gets its setup from `kickoff-worktree.sh`.

**After creating each additional worktree, run
`spectre link --root <abs-worktree>/spectre <canonical-peer>:<name>` with the working directory at
that repository's primary checkout.**
`<canonical-peer>` is the canonical repository's own name in that worktree's
`<project>/spectre/peers` file. Each link writes `link.md` on both sides, so each is followed by its
link commit before the next link runs (**Planning commits**,
`skills/flow-contracts/git-boundaries.md`) — never by `--force`. Each link commit is made in the
worktree holding that `link.md`, never the primary checkout the link ran from, behind
`check-planning-commit-location.sh <abs-worktree> <name>`; a refusal stops the run like a refused
link. Record what the command wrote alongside that worktree's merge
base in this run's working notes. **A refusal is a hard failure of this stage**: report it and
stop the run. A change with no linked peers runs
nothing here.

**2. Isolate the workspace** (`skills/flow/implement.md`) records the worktree of every `## apps`
entry whose repository already holds one, and creates nothing for it; for each remaining entry:

Every other entry gets one call,
`<agents repo>/scripts/kickoff-worktree.sh --from <this change's worktree> <that repo> <name>`,
whose header is canonical for the steps: the location guard; resuming `origin/spectre/<name>`, or
cutting from `origin/<base>` when this change's worktree records a flowBase, else
`origin/<default-branch>`; the flowBase recorded in that repository; the merge base persisted into
this change's state record; that repository's `## worktree setup`; the push per **Branch backup**
(`skills/flow-contracts/git-boundaries.md`). No `spectre link` runs for these worktrees — they
are declared apps, not peers — and each of them joins this run's resolved worktree set. **A
worktree add that fails is a hard failure of this stage**, reported and stopping the run exactly
like a refused link. Its exit 1 naming a missing `origin/<base>` stops the run and asks the
operator the two remedies the message names: push `<base>` to that origin, or re-run naming that
repository's default branch as `<base>`.
The worktree of the repository `## visual verification`'s `regression checkout` names,
once created — or resumed without a `node_modules` — gets its own toolchain: that section's
`setup` command runs from the worktree's root, in the foreground, and a non-zero exit is a hard
failure of this stage like a failed worktree add. Never a symlink to the main checkout's
`node_modules`.

**Then make the merge-order record cover the whole set.** The canonical `link.md`'s
`## Merge order` is what **Finish contract** (`skills/flow-contracts/finish-contract-run1.md`)
reads to sequence run 1's routes, so every repository of the resolved worktree set — linked or
not — is named in it before this stage ends. When `spectre link` wrote the canonical side's
`link.md`, extend its `## Merge order` with the resolved repositories it does not yet name,
appended in `## apps` declaration order after the entries it found — never reordering what the
link wrote; when no link ran and the resolved set holds more than one repository, write the
change's own `<changeRoot>link.md` carrying a `## Merge order` section alone — never a
`## Part of`, the section that makes a `link.md` a satellite's — ordered `.` first, then the
remaining repositories in `## apps` declaration order. Every entry other than `.` is the
repository's name in the canonical worktree's `<project>/spectre/peers` file — the name `spectre
validate` accepts there — never the `## apps` row's label; a resolved repository that file does not
declare is a hard failure of this stage, reported and stopping the run like a refused link, until the
project declares it. The write lands as a planning commit in the
worktree holding the file, behind `check-planning-commit-location.sh <abs-worktree> <name>`
(**Planning commits**, `skills/flow-contracts/git-boundaries.md`), and what it wrote is recorded
in the working notes beside the merge bases. A change with one repository runs nothing here.

**Each `## Merge order` entry is a numbered line naming its repository in backticks** — `spectre
validate` refuses a bare `1. .` — so a change spanning this repository and one peer writes, per
**Finish contract** (`skills/flow-contracts/finish-contract-run1.md`):

~~~markdown
## Merge order

1. `.`
2. `<peer>`
~~~
