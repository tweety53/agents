# Isolate the workspace — additional worktrees

Loaded during `flow.isolate-workspace` by the load directive of **2. Isolate the workspace**
(`skills/flow/implement.md`), only when the change links a peer, an `## apps` entry names a
repository holding no worktree for this change, `## visual verification` names a
`regression checkout`, or the resolved worktree set spans more than one repository.

**Every worktree this stage creates — a linked peer's, or an `## apps` entry's below — runs its
repository's `## worktree setup` before anything else touches it**: before its `spectre link`, its
link or planning commit, and any task. Run `project-get.sh <that worktree> "worktree setup"` and
handle it exactly as step 4 of **A. Resolve the change and write `STARTED`**
(`skills/flow/brainstorm.md`) does for the kickoff worktree — only the fenced command lines, from
that worktree's root, in order, in the foreground; exit 1 says the key is absent and continues;
exit 2 stops the run; a command's non-zero exit ends the turn naming the command and its output. A
worktree this stage resumes rather than creates runs nothing here.

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

Every other entry gets the kickoff recipe in its own
repository — `<project>` in the commands below is that entry's repository root, not this
change's — in the form that run's state calls for (`skills/flow/brainstorm.md` step 3): when
`git -C <that repo> rev-parse -q --verify origin/spectre/<name>` succeeds — the branch an
earlier run pushed — `git worktree add <project>/.worktrees/<name> spectre/<name>`, a local
branch tracking that remote one; otherwise `git worktree add <project>/.worktrees/<name> -b
spectre/<name> origin/<default-branch>` — or `origin/<base>` when
`git -C <this change's worktree> config --get branch.<branch>.flowBase` prints a `<base>`,
`<branch>` being `spectre/<name>`, which is then recorded in that repository too, by
`git -C <that worktree> config branch.<branch>.flowBase <base>` — after
`check-worktree-location.sh <that repo>`, with
the merge base
`git -C <that worktree> rev-parse HEAD` prints persisted into the state file's `worktrees`
map by `flow state add-worktree <name> <that worktree> <merge-base>`, and
`git -C <that worktree> push -u origin spectre/<name>` per **Branch backup**
(`skills/flow-contracts/git-boundaries.md`). No `spectre link` runs for these worktrees — they
are declared apps, not peers — and each of them joins this run's resolved worktree set. **A
worktree add that fails is a hard failure of this stage**, reported and stopping the run exactly
like a refused link. The worktree of the repository `## visual verification`'s `regression checkout` names,
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
remaining repositories in `## apps` declaration order. The write lands as a planning commit in the
worktree holding the file, behind `check-planning-commit-location.sh <abs-worktree> <name>`
(**Planning commits**, `skills/flow-contracts/git-boundaries.md`), and what it wrote is recorded
in the working notes beside the merge bases. A change with one repository runs nothing here.
