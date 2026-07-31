# myflow pipeline

The three-state pipeline itself: state definitions, the command→state transition table, git
boundaries, the handoff shape, and the finish contract.

**Load this file when running any `/myflow-*` command.** It is split out of
`rules/myflow-manual-review.mdc` so the always-on rule layer carries only the trigger, not the
whole state machine — the same reason the four contract files beside it were split out.

This file is **canonical** for everything in it. Where a skill or command disagrees with it, this
file wins.

## States

A change is always in exactly one of three states, recorded in its state file.

```text
/myflow-start  → STARTED      you: read the proposal artifact
/myflow-do     → IN_PROGRESS  you: review the staged diff and run the apps
/myflow-finish → FINISHED     terminal (second run — see the finish contract)
```

**Each command ends in the state named after it**, so the state vocabulary and the command
vocabulary are the same words.

**The human gate is a property of the state, not a separate stage.** `IN_PROGRESS` *means* a
staged diff and a test guide are waiting for the human. Nothing records that the review or the
testing happened; the operator running the next command is what carries the change forward. This
is why no `*-done` command exists — there would be nothing for one to write.

| State | Means | Waiting on |
|-------|-------|-----------|
| `STARTED` | The proposal exists and is published | you — read the artifact |
| `IN_PROGRESS` | The implementation is staged and a test guide exists | you — review the diff, run the apps |
| `FINISHED` | Archived, pushed, worktrees removed | — |

**Reviewing and testing are one gate.** `/myflow-do` produces both surfaces in the same run, so
the human does both at one sitting. There is no state between implementation and finishing —
integration is not a stage, it is the first half of finishing.

## Command surface

Three pipeline commands and two read-only ones. **No command accepts a flag.** The only argument
is the optional change name — see **Change name resolution**.

Behaviour a flag used to select is now either asked at invocation (the integration choice in
`/myflow-finish`), derived from the current state, or fixed at the single sensible default
(review-panel breadth, decided by its escalation triggers).

An argument that is not a known change name is **reported**, not silently ignored — a silently
ignored word is indistinguishable from a flag that stopped working.

## State transitions

| Command | Accepts | Ends at |
|---------|---------|---------|
| `/myflow-start` | *(none — creates the change)* · `STARTED` | `STARTED` |
| `/myflow-do` | `STARTED` · `IN_PROGRESS` | `IN_PROGRESS` from `STARTED`; **otherwise unchanged** |
| `/myflow-finish` | `IN_PROGRESS` | `IN_PROGRESS` after run 1; `FINISHED` after run 2 |
| `/myflow-status`, `/myflow-info` | any — read-only, never block | unchanged |

**This table is authoritative.** Every command file — in **both** command trees (`commands/` and
`commands-claude/`) — must state exactly the states its row lists, and must agree with the skill it
delegates to. When a command and its skill disagree, whichever the agent reads first wins, which is
non-determinism in the one layer that must be deterministic.

### Every command is re-entrant

Re-invoking a command is the supported way to revise its output. There is no separate `*-fix`
command:

- **`/myflow-start` at `STARTED`** revises the proposal and republishes the artifact to the
  **same** URL.
- **`/myflow-do` at `IN_PROGRESS`** resumes the existing worktree and applies a fix, documenting it
  in `proposal.md`/`tasks.md` or a nested `<name>-fix-N` sub-change first, and refreshing the test
  guide alongside the code so the two surfaces never drift apart.
- **`/myflow-finish` at `IN_PROGRESS`** integrates on its first run and archives on its second.

### A fix never moves the state

`/myflow-do` advances the state **only** from `STARTED` to `IN_PROGRESS`. From `IN_PROGRESS` it
writes the state back exactly as it found it.

No field records where a fix was raised. Whether the human re-reviews the diff or re-runs the apps
after a fix is their decision.

## Wrong state for this command

**On a mismatch, stop.** Report the actual state, the states the command expects, and the command
that should run instead, then **AskUserQuestion** for an explicit override with **"No — run the
suggested command instead"** as the default and recommended answer. Only proceed when the user
explicitly chooses to override. Never advance from a wrong starting state silently.

```
## Wrong state for this command

**Change:** <name>
**Current state:** <actual> (set by <updatedBy>, <updatedAt>)
**This command expects:** <expected>
**Suggested instead:** /myflow-<other> <name>
```

## Git boundaries

| Command | Condition | Allowed git actions |
|---------|-----------|---------------------|
| `/myflow-start` | — | None — planning artifacts only |
| `/myflow-do` | from `STARTED` | Create branch/worktree + **`git add` excluding the planning paths** — no commits, push, merge, or PR |
| `/myflow-do` | at `IN_PROGRESS`, no `prUrl` | Resume **existing** worktree + **`git add` excluding the planning paths** — no commits |
| `/myflow-do` | at `IN_PROGRESS`, `prUrl` recorded | **Commits twice and pushes** to the PR branch — implementation, then planning artifacts; the one exception |
| `/myflow-finish` | run 1 | **Commits twice** — implementation, then planning artifacts — and pushes; opens a PR or merges, by the operator's choice |
| `/myflow-finish` | run 2 | **Commits and pushes the archive**; removes worktrees and branches |
| `/myflow-status`, `/myflow-info` | — | None — read-only |

**Why `/myflow-do` commits once a PR exists and not before.** A PR is a remote surface, so a
staged-only fix would be invisible on it. Before that the operator reviews a staged diff in the
IDE, and committing would take that away.

**The planning paths** are the three that
**Handoff output** (`pipeline.md`) names below. `/myflow-do` clears them from the index and only
then stages with them excluded by pathspec — an exclusion governs what an
`add` adds and cannot retract what an earlier step staged, so the clearing pass is what makes the
rule hold rather than merely assert it. Its staging area therefore carries implementation only, and
`/myflow-finish` is what commits them.

`git add -A` respects `.gitignore`. Never force-add.

**Both commits are guarded, and an empty one is skipped rather than failed.** `git commit` exits
non-zero when nothing is staged, so an unguarded two-commit sequence dead-ends on three ordinary
cases: a fix touching only the three planning paths leaves the implementation commit empty — which
is exactly what run 1's unfinished-work **Stop** course invites — a fix touching only implementation
leaves the planning commit empty, and a re-run after a rejected push finds both commits already
made. Each commit is therefore preceded by a staged-changes test, and the whole sequence is one
`&&` chain:

```bash
git -C <abs-worktree> reset -q -- openspec/ docs/manual-test/ docs/superpowers/ \
  && git -C <abs-worktree> add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/' \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "<type>(<name>): <what the implementation does>"; } \
  && git -C <abs-worktree> add -A \
  && { git -C <abs-worktree> diff --cached --quiet \
       || git -C <abs-worktree> commit -m "chore(<name>): plan, test guide and session records"; }
```

**A skipped commit is reported, and a FAILED commit stops the sequence.** Those are different
outcomes: "nothing to commit" is normal and costs one line in the handoff, while a commit a hook
rejects is a git failure and gets the standard treatment — report git's own output and stop. The
chain is what enforces the second, and it is a chain rather than `set -e` deliberately: bash before
4.0 ignores `set -e` inside a subshell whose parent has errexit off, and this block runs through an
agent's shell whose state it does not control. Run it as one command. Without that, a first commit
a hook rejects falls through to the unconstrained second `add`, and the sole resulting commit —
titled `chore(...)` — carries the implementation, silently breaking the very split this section
exists to enforce.

**A planning path that is a tracked symlink stops the run, and is never worked around.** When any of
the three is a symlink, `git add -A -- . ':(exclude)docs/superpowers/'` exits 128 with
`fatal: pathspec … is beyond a symbolic link` and stages **nothing at all**. Report that message,
name the path, and stop at `IN_PROGRESS`. The only way to stage past it is a bare `git add -A`,
which puts the planning artifacts into the implementation commit — the one outcome this split
exists to prevent — so the fix belongs in the repository, by making the path a real directory.

## Handoff output

Every command ends in the same shape, and prints **nothing** after it:

```
<1–3 lines: what actually happened>

<absolute paths to anything the operator needs to open>

Next:
/myflow-finish <name>
```

- **The next command is the last line** — bare, copy-pasteable, with no prose after it. An agent
  cannot drive a harness's autocomplete; nothing lets a running session prefill the operator's
  input box. The last-line convention plus a five-command surface is the whole mechanism.
- **`/myflow-finish` run 1 names itself** as the next command, because that is what the operator
  runs once the branch is merged. Only a run 2 that **completed** is terminal and names nothing — a
  run 2 that stopped on a cleanup leftover names itself too, for the same reason: the operator
  clears what remains and runs it again.
- **Only what the operator must act on.** Do not restate the plan, enumerate completed internal
  steps, or repeat content available at a path you just gave.
- **Link, never paste.** Manual test guides, diffs and plans are given as absolute paths.
- **Every path is absolute** — in handoffs, in generated guides, in IntelliJ commands, in run
  instructions. Never a relative path, never `../<other-app>`, and never a main-checkout path while
  an apply worktree holds the work. Resolve app roots from `git worktree list` or the state file's
  `worktrees` keys.
- **`openspec/`, `docs/manual-test/` and `docs/superpowers/` are never staged before finish.** The
  plan was read at `STARTED`; presenting it again as code to review hides the implementation diff it
  is mixed into. Leaving them unstaged, rather than filtering them out of one display command, is
  what makes them absent from *every* view of the staging area — a filtered display leaves them in
  the index, where `git status`, a graphical client and the IDE's staged-changes pane show them
  again. The list is fixed here rather than configured per project; the pipeline chooses these paths
  itself, so no project can differ. `/myflow-finish` run 1 stages them and commits them separately
  from the implementation, so nothing is lost:

  ```bash
  git -C <abs-worktree> diff --cached --stat
  git -C <abs-worktree> diff --cached
  ```

## IntelliJ commands

Every state that waits on a human must print a copy-paste command in its handoff:

```bash
open -na "IntelliJ IDEA" --args "<absolute path>"
```

Use `open -na`, not the `idea` shim — that shim is not on this machine's PATH. `open` resolves the
app by name (bundle `com.jetbrains.intellij`), returns immediately instead of blocking the shell,
and reuses a running instance.

| State | Path to open |
|-------|--------------|
| `STARTED` | main checkout (artifacts live there; no worktree exists yet) |
| `IN_PROGRESS` | apply worktree root, plus the test guide's absolute path |

Paths are absolute, resolved from `git worktree list`. Never emit a relative path.

## Finish contract

`/myflow-finish` is a **two-run** command. Which run happens is decided by one thing: whether the
change's branch has already reached the base branch. No field records "integration started" — the
branch's merge status is the only source of truth, and a field could disagree with it.

**The merge status is decided by three signals, in this order, and by a script — not by prose.**
`scripts/check-finish-preflight.sh <worktree> <base-ref> <recorded-merge-base|->` prints one verdict
line and exits 0 whenever it reached a verdict. It exits 2 with no verdict when it cannot read the
worktree at all — an unreadable tree is never a licence to proceed.

| Verdict | Meaning |
|---------|---------|
| `RUN1` | integrate — the branch has not reached the base branch |
| `RUN2` | archive — merged, and nothing is outstanding |
| `REFUSE` | stop and ask the operator before anything is archived |

1. **`HEAD` against the merge base recorded in the state file's `worktrees` map.** No recorded
   value, or one that does not resolve → `REFUSE`; an honest unknown is never inferred. Equal to
   `HEAD` means the branch has no commits of its own → `RUN1`, whatever the ancestor test says.
   Both are answered before the base ref is resolved, so no environmental failure can hide them.
2. **The ancestor test** — `git merge-base --is-ancestor`, and only that. The script consults no PR
   CLI: it must give the same answer on every forge, and git alone already answers this question.
   The base ref is resolved first, so a base ref that does not resolve is its own `REFUSE` rather
   than an accidental `RUN1`.
3. **The worktree's cleanliness.** Merged by ancestry with uncommitted entries → `REFUSE`.
   Merged, distinct from the recorded merge base, and clean → `RUN2`.

**Signal 1 precedes signal 2, and that ordering is the point.** A branch with no commits of its own
is an ancestor of every branch, so the ancestor test alone reports *merged* on a branch whose work is
staged and never committed — after which run 2 archives the change and `--force`-removes the worktree
holding all of it.

**Never substitute a commit count.** `git rev-list --count <base>..HEAD` is zero both for a branch
with no commits and for a branch whose commits have joined the base branch, so it cannot separate the
dangerous state from the correct terminal one, and using it would refuse every legitimate archive.

On a `REFUSE`, stop before touching anything, report `HEAD`, the base branch and the uncommitted
count, and ask the operator explicitly. On a multi-repo change, run the script once per `worktrees`
key and proceed to run 2 only when **every** worktree returns `RUN2`.

**When the script is absent** — a harness whose repository does not carry it — perform the same three
signals by hand in the same order and say in the handoff that the check was run manually. The check is
never skipped for want of the script. Signal 2 may then be answered by a PR CLI when one is usable
for the host, as in run 2 below — that option belongs to the human doing this by hand, never to the
script — but signals 1 and 3 still run, and still run in this order.

### Run 1 — the branch is not merged

**Check for unfinished work first — before the landing question and before any git action.**
`scripts/check-unfinished-work.sh <worktree> <change-name>` prints one verdict line and exits 0
whenever it reached a verdict. It exits 2 with **no** verdict line when it cannot read the worktree.
Run it once per worktree recorded in the state file's `worktrees` map.

| Verdict | Meaning |
|---------|---------|
| `CLEAR` | nothing outstanding — go straight to the landing question, with no extra prompt |
| `OUTSTANDING` | show the breakdown and offer the three courses below |

A missing verdict line is not a verdict. Treat it exactly as the preflight script's fourth outcome
above: stop and ask the operator. The exit code is checked as well as the line, because a caller
that greps for `CLEAR` in empty output finds nothing.

On `OUTSTANDING` the operator is offered **exactly three** courses:

| Course | What run 1 then does |
|--------|----------------------|
| **Stop — I'll finish it first** *(recommended)* | stop, leaving the change at `IN_PROGRESS` with nothing staged, committed or pushed |
| **Continue — integrate anyway** | proceed to the landing question, carrying the outstanding list into the planning commit's message and the handoff |
| **File a Jira task, then continue** | file an issue carrying the outstanding items, then proceed |

**Stop is marked as the recommendation, and the reason is stated rather than left to be inferred.**
The gate only fires because something really is unfinished, and finishing it is the cheapest of the
three to recover from — Continue is the only course that reaches an irreversible step, and it exists
for work the operator deliberately deferred, which is a judgment only they hold. Marking a
recommendation is not a courtesy here: the planning-gate capability requires every choice a
`/myflow-*` command offers to name its recommended option, and this prompt is one of them.

There is no fourth course, and in particular none that hands back to `/myflow-do` inline. The filed
issue is labelled and linked per
**Labels on issues the pipeline creates** (`skills/myflow-contracts/jira-integration.md`).

**A filing that fails is one skipped-with-reason line, and the run still proceeds** — the same
degradation every other Jira write in this pipeline has, per
**Never blocking** (`skills/myflow-contracts/jira-integration.md`). Creation can fail for the usual
reasons (auth, permission, an unknown project key or label) and there may be no tracker configured
at all. None of them changes the operator's answer, which was *continue*: the outstanding list still
reaches the planning commit's message and the handoff, which is where this change requires the
durable record to be. A failed filing is never silently upgraded to **Stop**, and never passes
unmentioned.

**The gate precedes the question, and that ordering is the point.** An operator asked how to land a
branch, and only then told it carries unfinished work, has already answered a question about a
branch they believed was complete — and the cheapest of the three courses to take by mistake is the
one that integrates.

**What the operator integrated over is recorded where a transcript is not**: the outstanding list
goes into the message of the commit that carries the planning artifacts, and into run 1's handoff.
The signals that produce that list are the script's own and are deliberately not restated here.

Only then ask, **before any git action**, how the branch should land:

> **How should this branch land?**
> - **Open a pull request** *(default, recommended)*
> - **Merge and push**
> - **Handle it manually**

Then run to completion without asking again. The answer is never remembered between runs.

All three routes first commit the work, in **two** commits and never one: the implementation, then
`docs/manual-test/<name>.md`, the `openspec/` planning artifacts and the session records preserved
under `docs/superpowers/` (the SDD ledger, the review panel record, and the proposal artifact
source). The records are copied out of the gitignored worktree before staging, by
`scripts/preserve-session-records.sh <worktree> <name> <state-dir>`.

Those three planning paths are cleared from the index before the first `add` and excluded from it by
pathspec — the same clearing pass **Git boundaries** (`pipeline.md`) above gives `/myflow-do`, and
for the same reason: an exclusion cannot retract what an earlier step staged, and at this gate that
step may have been the operator's own `git add`. The second `add` carries no pathspec, which is what
picks the three paths up. The sequence itself — the guarded commits, the skipped-empty rule, the
failure rule and the symlink case — is the chain **Git boundaries** (`pipeline.md`) gives, and is
not written out a second time here.

**Implementation first is the order, not an accident of it.** The newest commit is the one a forge
shows first, and that should be the code; and the second commit's message is where the outstanding
list from the gate above is written down.

**That script has three outcomes, and they are not interchangeable.** This is where they are
defined; the two call sites point here rather than each describing them.

| Outcome | What it means | What you do |
|---------|---------------|-------------|
| `skipped: <src> (absent)`, exit 0 | The source does not exist. A change may legitimately have no panel record. | Nothing. Proceed. |
| `preserved: <dest>`, exit 0 | The record reached the repository at that path. | Nothing. Proceed. |
| A message on stderr, **exit non-zero** | A copy was attempted and refused or failed — an untrusted source or destination path, or a write that could not be made. | **Report it to the operator, with the script's own message.** Then proceed with the integration. |

**A non-zero exit is never silent and never a stop.** Those two rules pull in opposite directions
and both hold: a preservation step able to abandon an integration whose work is already committed
would be a worse failure than the missing record, *and* a refused write that passed unmentioned
would leave the operator believing a record exists when none does. So it is reported and the run
continues — and the handoff says which records were preserved and which were not. The remaining
sources are still attempted after any one failure, so a single bad path costs one record, not three.

**That ordering is deliberate, and deliberately differs from `/myflow-do`'s.** Here the preservation
call comes *before* any staging, because all three routes commit: the second, unconstrained `add`
covers the preserved files wherever in the run they were written, and placing the call first keeps it
out of the route branches, where it could be forgotten on one of them. `/myflow-do` runs the same
script *after* its unconditional staging and stages a second time, because it stages on every run
and commits only when a `prUrl` is already recorded; hoisting the call there would create
`docs/superpowers/` files on every ordinary staged-only run.
**Do not harmonise the two orderings for symmetry** — the asymmetry is what keeps preserved records
out of the staged-only path.

| Route | Then |
|-------|------|
| **Open a pull request** | push; open a PR via `gh` when usable for the host, else print the forge's create-PR URL and ask whether it was opened; record `prUrl` |
| **Merge and push** | push; merge into the base branch; push that |
| **Handle it manually** | push the branch only; say plainly what is left to do |

Run 1 ends at `IN_PROGRESS`, and its handoff's last line is `/myflow-finish <name>` again.

**Resolve the base branch; never assume it, and never derive it from the current branch.**

```bash
# Both network calls are wrapped: `git remote show origin` against an unreachable host blocks for
# ~75s on the default TCP timeout, which would turn a correct refusal into a two-minute hang.
git -c core.askpass=true fetch --quiet origin 2>/dev/null || true   # a stale ref only fails safe
BASE="$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's#^origin/##')"
if [[ -z "$BASE" ]]; then
  BASE="$(GIT_TERMINAL_PROMPT=0 git remote show origin 2>/dev/null | sed -n 's/^ *HEAD branch: //p')"
fi
CUR="$(git branch --show-current)"
[[ -n "$CUR" ]] \
  || { echo "detached HEAD — check out openspec/<name> before finishing"; exit 1; }
[[ -n "$BASE" ]] \
  || { echo "no base branch resolved. If this repo has no remote, finish cannot integrate — see below"; exit 1; }
[[ "$BASE" != "$CUR" ]] \
  || { echo "base branch resolved to the current branch ($CUR) — refusing to compare it with itself"; exit 1; }
```

**Never fall back to `HEAD@{upstream}`.** `/myflow-finish` runs inside the apply worktree, where
`HEAD` *is* `openspec/<name>` — so that fallback resolves to the change's **own** upstream, making
the merge check `openspec/<name>` vs `origin/openspec/<name>`, which is true the moment the branch
is pushed. That silently reports an unmerged change as merged, and run 2 then archives it and
deletes its worktree. Asserting `BASE` differs from the current branch is what makes that class of
misresolution impossible rather than merely unlikely.

If no base branch resolves, **stop and ask**. An unresolvable base is an honest unknown; a guessed
one is a wrong answer at the only irreversible step.

**A repository with no remote at all cannot be integrated by this command.** Every route needs a
push, and base resolution needs `origin`. Say exactly that — *"this repository has no remote, so
there is nothing to push to or merge into"* — rather than reporting a base-branch failure, which
sends the operator debugging the wrong thing. Offer to leave the change at `IN_PROGRESS` with the
work staged; there is nothing to lose, because nothing was pushed.

**No verification gate runs before integration.** No tests, no linters, no spec-coverage check.
Correctness was established during `/myflow-do` — TDD per task, per-task review, the final review
panel — and by the human gate. Re-running it here would repeat finished work immediately before
the one irreversible step.

### Run 2 — the branch is merged

1. **Verify the merge.** Use a PR CLI when one is usable for the host; otherwise
   `git merge-base --is-ancestor`. That fallback must stay reachable on its own — it is the only
   merge evidence available on a non-GitHub forge. **Not merged → this is not run 2.**
2. **Sync delta specs** into `openspec/specs/`, then move the change into
   `openspec/changes/archive/<date>-<name>/`. Any nested `<name>-fix-N` sub-changes are archived in
   the same operation — never left behind, never archived alone.
3. **Commit and push the archive** on the base branch in the main checkout. There is no merge to do
   in the normal case: the change branch was already merged, which step 1 proved. When finish is
   invoked with a non-base branch checked out, it commits there, merges into the base branch, and
   pushes that. A finished change never leaves the archive move uncommitted in the working tree.
4. **Clean up the worktrees, the local branch and the remote branch** — see
   **Worktree cleanup** (`pipeline.md`) below.
5. **Remove the proposal artifact source** from the state directory, on the condition its row in
   **Temporary artifacts registry** (`pipeline.md`) gives. That section carries the condition and
   the reason for it; this step does not repeat either.
6. **Verify the cleanup.** Run `scripts/check-cleanup-complete.sh <repo> <name> <state-dir>` once
   per repository, **after** every removal above — it is there to judge what the run actually left
   behind, which is the one thing run 2 previously assumed.

   | Verdict | What run 2 does |
   |---------|-----------------|
   | `COMPLETE:` | report the cleanup as verified, and go on to step 7 |
   | `LEFTOVER:` | name what remains, **do not write `FINISHED`**, and stop at `IN_PROGRESS` |

   A non-zero exit with **no verdict line** is the third outcome and not a verdict: report it, leave
   the affected `worktrees` entries in the state file, and treat it exactly as `LEFTOVER` — an
   unverified cleanup is not a verified one. The exit code is checked as well as the line, because a
   caller that greps for `COMPLETE` in empty output finds nothing.

   **A leftover blocks the `FINISHED` write, and that is the whole point of having a verdict.**
   `FINISHED` is terminal: `/myflow-finish` stops at it and `/myflow-status` does not list it, so a
   change written `FINISHED` over a known leftover has exactly one record of that leftover — the
   console line — which is the transcript-only record this pipeline refuses everywhere else. Left at
   `IN_PROGRESS` instead, the change stays listed, stays re-runnable, and the state file it already
   has is the durable record; no new field is invented to carry a fact the state itself carries.

   **Run 2 is re-entrant, which is what makes that safe.** Every step is remove-or-move *if present*
   and a step whose artifact is already gone is success, not an error — so a re-run after the
   operator clears the leftover repeats the verification and nothing else. An already-archived change
   directory means step 2 is already done: sync and archive are skipped, not repeated, and the run
   continues to cleanup and verification.

   **When the script is absent** — a repository that does not carry it — check the same registry rows
   by hand, in the same order, and say in the handoff that the verification was done manually. The
   check is never skipped for want of the script, and "not verified" is never reported as verified.
7. **Write `FINISHED`**, clearing from `worktrees` **only the entries whose removal actually
   succeeded** — see **Worktree cleanup** (`pipeline.md`) below — and carry every other field
   forward. This step is reached only on `COMPLETE:`.

### Worktree cleanup

The set of worktrees is the **keys of the state file's `worktrees` map**. When that is absent, scan
each affected repository:

```bash
git -C "$REPO" worktree list --porcelain \
  | awk '/^worktree /{w=substr($0, 10)} /^branch /{if ($2=="refs/heads/openspec/<name>") print w}'
```

**Never guess a path.** Worktree layout differs per repository — this repo keeps its worktrees in a
sibling directory, not `.worktrees/`.

**The path is taken with `substr`, never `$2`.** `worktree list --porcelain` emits it raw, so a
field reference truncates any path containing a space at the first one: fed
`worktree /tmp/my worktree/x` it yields `/tmp/my`, and the run then `--force`-removes a path that is
not the worktree, or fails having named the wrong one. `10` is one past the length of the literal
`worktree ` prefix. The branch on the next line is a ref name and cannot contain a space, so `$2` is
right for it. `scripts/check-cleanup-complete.sh` parses the same stream the same way — the guard
and the snippet it verifies must not disagree, or the wrong one gets copied next.

For each worktree, run **all four** checks before removing anything:

```bash
# 1. no uncommitted tracked changes — must be empty
git -C "$WT" status --porcelain --untracked-files=no

# 2. no untracked files that git does not already ignore — must be empty.
#    `--others --exclude-standard` lists exactly the files `--force` would destroy and
#    `.gitignore` does NOT cover. This is the check that makes `--force` safe.
git -C "$WT" ls-files --others --exclude-standard

# 3. no commits that exist only here. `@{upstream}` ERRORS when no upstream is configured, and
#    an empty capture would read as "nothing unpushed" — so resolve it explicitly and never let
#    a failed lookup pass as success.
#
#    Step 1 already proved the branch is an ancestor of the base branch, which is STRICTLY
#    STRONGER evidence than "pushed to its own upstream": the commits are in the base branch.
#    Accept that first. Requiring the upstream regardless would lock out the ordinary
#    squash-merge workflow — GitHub's "delete head branch on merge" plus `fetch.prune=true`
#    removes the tracking ref, after which no upstream can ever resolve and the branch cannot be
#    re-pushed because it no longer exists on the remote.
if git -C "$WT" merge-base --is-ancestor HEAD "origin/$BASE" 2>/dev/null; then
  :                                            # already merged into base — nothing can be lost
elif UP="$(git -C "$WT" rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null)"; then
  git -C "$WT" log --oneline "$UP..HEAD"       # must be empty
else
  echo "not merged, and no upstream — cannot prove these commits exist anywhere else"; false
fi

# 4. what `--force` WILL destroy: ignored files. This does not gate removal — it is shown to
#    the operator, who decides. `--exclude-standard` in check 2 hides everything matched by
#    .gitignore, .git/info/exclude or the global excludes file, and "ignored" is NOT "disposable":
#    a deliberately-ignored .env, a local override config, or this pipeline's own
#    .superpowers/sdd/ records are all ignored and all irreplaceable.
git -C "$WT" ls-files --others --ignored --exclude-standard

# 5. the project's local stack is stopped — run its `## stop` command if declared. Give it a
#    bounded wait — 60 seconds — and treat a timeout as a FAILED check.
```

**Check 4 is a disclosure, not a gate.** When it lists anything, **stop and show the list**, and
ask for explicit confirmation before removing that worktree. Do not try to classify the entries as
build output: no allowlist of names can be trusted, because the operator decides what they ignore.
Empty list → proceed without asking.

Then, and only then:

```bash
git -C "$REPO" worktree remove --force "$WT"
git -C "$REPO" branch -d "openspec/<name>"
git -C "$REPO" worktree prune
```

- **`--force` destroys every ignored file in the worktree, and no check prevents that.** Checks 1
  and 2 establish only that nothing *tracked-and-modified* and nothing *untracked-and-unignored*
  is at risk. They say nothing about ignored files, because `--exclude-standard` is what hides
  them — and "ignored" is not "disposable". Check 4 exists to make that visible rather than to
  prevent it: it lists exactly what will die, and the operator confirms. Claiming the checks make
  `--force` safe would be false, and was: a gitignored `.env` passes checks 1 and 2 and is
  destroyed silently.
- **Neither check sees a file whose `assume-unchanged` bit is set.** `git status` is blind to it
  by design. Rare, operator-inflicted, and named here so it is a known limit rather than a
  surprise.
- **`git branch -d`, never `-D`.** It must be free to refuse an unmerged branch.
- **An already-removed worktree is success**, not an error.
- **Any failed check leaves every worktree alone** and reports why. There is no partial cleanup.
  This includes check 4: a `## stop` command that **exits non-zero, is not found, or has to be
  interrupted** is a *failed* check, not an absent one — only an undeclared key is skipped. Give it
  a bounded wait rather than letting it hang the run.
- **Verify each removal actually succeeded** before writing state. If any `git worktree remove`
  fails for a reason the checks did not predict — a file lock, a permission error — report it and
  leave that worktree's entry in `worktrees`. Writing `worktrees: {}` regardless would drop it from
  the only authoritative list, and nothing would ever find it again.

Then the change's **remote** branch:

```bash
# `push --delete` exits non-zero BOTH when the branch was already gone and when the push was
# refused, so the two are told apart by git's message and never by the exit code alone. Measured
# against a scratch remote on this machine (git 2.50.1): an already-absent branch prints
# `error: unable to delete 'openspec/<name>': remote ref does not exist` and exits 1, and the
# stale remote-tracking ref SURVIVES that failure.
OUT="$(git -C "$REPO" push origin --delete "openspec/<name>" 2>&1)"; RC=$?
if [ "$RC" -eq 0 ]; then
  echo "remote branch deleted: origin/openspec/<name>"
elif printf '%s' "$OUT" | grep -q 'remote ref does not exist'; then
  # The forge deleted it on merge. Prune the ref it left behind: check-cleanup-complete.sh reads a
  # surviving refs/remotes/origin/openspec/<name> as a leftover, and it would be a real one.
  git -C "$REPO" fetch --prune --quiet origin
  echo "remote branch already gone — the forge deleted it on merge"
else
  echo "remote branch NOT deleted: $OUT"
fi
```

- **The remote branch is deleted without a further prompt.** Run 2 is reached only by proving the
  branch is an ancestor of the base branch, so its commits are in the base branch and nothing can be
  lost — which is why this is not gated the way check 4's disclosure is.
- **An already-absent remote branch is success**, not an error, and the outcome is reported either
  way: deleted, already gone, or refused.
- **A refused push is reported, never swallowed.** A bare `|| true` would make an expired
  credential indistinguishable from a branch the forge already removed, and leave the remote branch
  standing with nothing said about it.
- **The remote delete is not gated on the local one succeeding.** Gating it would leave the remote
  branch behind whenever anything unrelated failed, which is the state this step exists to end.

The stack-stopped check reads the optional `## stop` key from the project's `.myflow/project.md` —
see **Project configuration** in `skills/myflow-contracts/project-configuration.md`. When the key
or the file is absent the check is **skipped, not failed**, and cleanup proceeds on the strength of
the other two.

## Temporary artifacts registry

Every artifact the pipeline creates, with what creates it, where it lives, and what removes it.

| Artifact | Created by | Location | Removed by |
|----------|-----------|----------|-----------|
| Per-task and review diffs | `/myflow-do` | `.superpowers/sdd/` in the worktree | with the worktree, at run 2 |
| Panel record | `/myflow-do` | `.superpowers/sdd/` | preserved at run 1; removed with the worktree |
| SDD ledger | `/myflow-do` | `.superpowers/sdd/` | preserved at run 1; removed with the worktree |
| Proposal artifact source | `/myflow-start` | the state directory | run 2, only if a preserved copy exists |
| Worktree | `/myflow-do` | per the `worktrees` keys | run 2, after its existing checks |
| Local branch | `/myflow-do` | the repository | run 2, `git branch -d` |
| Remote branch | finish run 1 | `origin` | run 2, without a further prompt |
| Change directory | `/myflow-start` | `openspec/changes/<name>/` | moved to the archive, never deleted |
| State file | every command | the state directory | never — it is the terminal record |

**This table is the one place a cleanup rule is stated.** Everything else that mentions a removal
points here rather than restating it, in this file and in every skill.
**Worktree cleanup** (`pipeline.md`) above is the *procedure* for the rows removed there and not a
second statement of the rule: the table says what is removed and when, that section says how — and
a stale second copy of a rule governing
`git worktree remove --force` is a copy that deletes the wrong thing.

**An artifact no row accounts for is a defect in the registry**, corrected by adding the row. It is
never left unaccounted for on the grounds that something probably removes it — that assumption is
exactly how the remote branch went unremoved until it was given a row here.

**Where the proposal artifact source comes from, and why its row is conditional.** `/myflow-start`
writes `<state-dir>/<name>-proposal-artifact.html` so a revision round can republish to the same
URL, and the preserved copy its row requires lives under `docs/superpowers/artifacts/`. The terminal
state file keeps `artifactUrl` indefinitely, so deleting the only source that could republish that
URL would leave it advertised and unrepublishable. No preserved copy → leave the file and say so.
The deletion is disclosed the same way the worktree removal is.

**Which rows run 2 verifies is read off this table, not listed again.** Every row whose lifetime
ends at run 2 is checked back by `scripts/check-cleanup-complete.sh`, whose header explains which
rows that leaves it reading and why; step 6 of
**Run 2 — the branch is merged** (`pipeline.md`) above is where its verdict is acted on.

**That derivation is declared, not left implicit.** The guard carries one marker line per row of
this table saying whether it checks that row or deliberately does not, with the reason; its harness
reads this table and those markers and fails when the two disagree in either direction. So a row
added here goes nowhere until someone records a decision about it — which is what stops a future
artifact from being confirmed clean by a guard that never looked for it.

## State file

The contract governing where a change's state file lives, its full JSON shape, monotonic state
writes, and carry-forward rules. **State file** (`skills/myflow-contracts/state-file.md`) — load it
before reading or writing a state file.

## State self-heal

The contract governing how a state file is validated against on-disk artifacts, and how a missing,
unparseable, or contradicted file is corrected.
**State self-heal** (`skills/myflow-contracts/state-self-heal.md`) — load it before self-healing a
state file.

## Project configuration

The contract governing `.myflow/project.md` — its optional keys, how a `## standards` entry
resolves to a file, and the containment rules that keep resolution safe.
**Project configuration** (`skills/myflow-contracts/project-configuration.md`) — load it before
resolving project configuration.

## Jira integration

The contract governing how a change is linked to a Jira issue, transitioned through the pipeline,
and has its description synced — including that Jira is never a gate and never blocks a state
write. **Jira integration** (`skills/myflow-contracts/jira-integration.md`) — load it before any
Jira-related step.

## Model policy

`/myflow-start` should run on **Opus** (or the harness's strongest available model) — brainstorming
and design benefit most from stronger reasoning. Every other `/myflow-*` command should run on
**Sonnet** (or the harness's standard default), and **every review-panel reviewer runs on Sonnet**
regardless of the parent model.

**Implementer subagents dispatched by `/myflow-do` run on Opus** (or the harness's strongest
available model). This **explicitly overrides** superpowers:subagent-driven-development's model
guidance — that skill says to pick "the least powerful model that can handle each role" and to use
the cheapest tier where the plan already contains the code to write. That guidance does not govern
this pipeline. The saving it offers is false here: an implementation defect is not avoided by the
review panel, it is *found* by the panel, at the cost of a fix wave and a re-run of every slot.
Buying a cheaper implementer with a more expensive review is the wrong trade.

The two rules point in opposite directions on purpose. A reviewer's job is to be many independent
readings of a finished diff, so its cost must not scale with the operator's session model. An
implementer's job is to get the diff right the first time, where capability compounds.

**Two further instructions in that same upstream skill are also overridden, and are named here
rather than left to be discovered.** subagent-driven-development says to dispatch the *final
review* on the most capable model: myflow does not — it fixes every panel slot at Sonnet, for the
reason above, and escalates the panel's **breadth** instead (the conditional Security, Adversarial
and extra-principle-lens slots), which buys more independent readings rather than one stronger one.
It also says to *escalate the model in fix rounds 4-5*: myflow cannot, because its implementers
already sit at the ceiling from round 1. Fix rounds escalate the same way — more lenses, not a
bigger model — and round 5 hands back to the operator rather than pretending an escalation is
available.

**An explicit operator instruction overrides either default, in either direction** — raising the
panel to Opus for a change that warrants it, or lowering the implementer for genuinely mechanical
work. Record the instruction with the dispatch; an override nobody wrote down is indistinguishable
from a mistake.

**Every subagent dispatch records the model it used** in the SDD ledger, alongside the task it ran.
A model policy that nothing records is a policy nothing can verify — and the absence of that record
is precisely how this rule came to be missing in the first place.

Where the dispatcher **cannot know** the model, the ledger records `unknown (agent-defined)` and
never a guess. Slots dispatched by `subagent_type` (Bugbot, Security Review) resolve their model
from their own agent definition, which the dispatcher does not read; writing a plausible slug for
them puts an unmeasured value into the audit trail.

**This record outlives the change.** The ledger is authored under `.superpowers/`, which is
gitignored, in a worktree `/myflow-finish` run 2 removes — but run 1 preserves it into the
repository first, under `docs/superpowers/ledgers/`, so it serves the operator and the panel
*during* the change and stays answerable afterwards. An after-the-fact audit of which model
implemented which task therefore reads the preserved ledger rather than a transcript nobody kept.
The preservation duty itself is stated once, under
**Run 1 — the branch is not merged** (`pipeline.md`); this section depends on it rather than
restating it.

Durability is a **stronger** reason to leave an unobserved entry unobserved, not a weaker one. A
persisting record makes an invented model slug permanent, so `unknown (agent-defined)` stays exactly
as written above and no step fills it in on the way into the repository.

- **Claude Code**: the **session** model is enforced via `model: opus` / `model: sonnet` in each
  command's frontmatter (`commands-claude/*.md`) — no manual action needed *for the session*.
  Frontmatter cannot set a **subagent's** model, so the implementer rule above is not enforced by
  it: `/myflow-do` must name the model on each implementer dispatch, and the ledger line for that
  task is what records that it did. Slots dispatched by `subagent_type` (Bugbot, Security Review)
  carry their own agent definitions and take no override from either mechanism.
- **Cursor**: not enforceable yet (no per-command model frontmatter support as of this writing) —
  each `.cursor/commands/myflow-*.md` file carries an explicit note; switch models manually in the
  composer/chat picker.
- **Codex**: no per-command/skill model override mechanism either — model is a session or profile
  level setting; switch manually before starting a new proposal.

## Change name resolution (all `/myflow-*` commands)

`<name>` is **optional** on every `/myflow-*` command. When omitted:

- Run `openspec list --json` and filter to changes relevant to that command's state (not yet
  archived).
- Exactly one match → use it automatically; announce which change was picked.
- Multiple matches → **AskUserQuestion** listing each (name, state, last modified) — never guess.
- Zero matches → fall back to that command's normal "no change" handling (e.g. `/myflow-start` asks
  what to build; others suggest the prior state's command).

A change linked to a Jira issue is named `<lowercased-key>-<slug>` — see
**Change naming** in `skills/myflow-contracts/jira-integration.md`.
