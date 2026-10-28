# Verify, stage, and hand off

Loaded by `skills/flow/SKILL.md` once `skills/flow/review-panel.md` closes clean. Stage order:
**verify → visual-verify → stage-diff → run-instructions → write-in-progress**.

## Verify

**Load `skills/flow-contracts/worktree-resolution.md`** before resolving this run's worktree set,
below.

```bash
flow stage begin -command '/flow' \
  -stage flow.verify \
  -harness <harness> \
  -session-token mf-<literal-token> \
  <name>
```

**First, validate the section and export what it declares — with the script, not by eye.** Run

```bash
prepare-workspace.sh <worktree>
```

once per worktree in this run's resolved set — the same set **2. Isolate the workspace**
(`skills/flow/implement.md`) resolved, non-empty by construction — never a raw read of the state
file's `worktrees` map. Per **Resolving a change's worktrees**
(`skills/flow-contracts/worktree-resolution.md`), report an empty resolved set and do not proceed.

`prepare-workspace.sh` runs `check-workspace-isolation.sh` against the worktree first, then — only
if that passes — derives and exports the variables the project's `## workspace isolation` section
declares, resolved against the workspace id, and prints one `KEY=value` line per exported variable
to stdout. Exit 0 means the printed lines are what to carry forward into `## lint` and `## test`
below (nothing printed means the project declares no `## workspace isolation` section). A non-zero
exit is the dropped-row case (exit 1, relay the script's own lines verbatim and stop) or the
cannot-answer case (exit 2, stop the same way) — stop **before** `## lint` and `## test`, without
writing the state file.

**Load `skills/flow-contracts/workspace-isolation.md` by section, only the one each trigger
needs** — `grep -n` for its heading plus `sed -n` for that section, never the whole file: **The
cache index** when `prepare-workspace.sh` exited 0 with stderr naming a `cache index` row, **The
empty id** when it exited 1, and nothing when it exited 2; the ordinary exit-0 run loads nothing.

**A declared `cache index` row is never among the printed `KEY=value` lines** — the script reports
it by name on stderr instead. On an exit-0 run whose stderr names a `cache index` row, probe the
project's cache here, claim a free index atomically, and record that claim in the cache itself under
an entry naming this workspace, per **The cache index**
(`skills/flow-contracts/workspace-isolation.md`).

**When the script cannot be located**, apply the same rules by hand from **Project configuration — workspace isolation** (`skills/flow-contracts/project-configuration-isolation.md`)
and **Workspace isolation** (`skills/flow-contracts/workspace-isolation.md`), and say in the handoff that the validation and
export were performed manually and why.

**Load `skills/flow-contracts/project-configuration-isolation.md`** only when the script cannot be located.

**Then compare every recorded base with the branch it tracks, on the run's first `flow.verify` only.** Run

```bash
check-base-behind.sh <worktree>...
```

over the same resolved set; its header is canonical for its lines. Exit 0 → continue. Exit 1 → relay every `BASE-BEHIND:` block and run its two command lines in order — the first pushes the base, unasked (**What still stops**, `skills/flow-contracts/operator-prompts-auto-resolution.md`); on zsh write `${C}:refs/heads/…`, since a bare `$C:r…` is the `:r` modifier — then run **Verify** again from its `begin` mark. Exit 2 → name its stderr in the handoff and continue (kan-1007: a stale `--base` was papered over by merging develop into the change branches, and integrate's reshape folded KAN-1010 into the squash).

**This step does not call the project's `create` command.** `create` is called by whatever starts
the project's applications, per **Project configuration**
(`skills/flow-contracts/project-configuration.md`), and this step starts them only for **Live check** below, through the project's own `## run`.

**After the panel closes, the parent edits no source outside the in-run fix loop.** Source
changes only through **The loop** (`skills/flow/verify-fix-loop.md`) or a fix run the operator
starts; any other source change from here on makes every slot's result stale (**Panel re-runs**,
`skills/flow/review-panel.md`).

**Load `skills/flow/verify-fix-loop.md`** only when this stage's or `flow.visual-verify`'s final
report carries a fixable defect — it carries **The loop** and **No cap**.

### Inline verify

**A worktree with nothing outside its change directory runs none of the list below.** In a run
whose resolved set holds more than one worktree, a worktree is skipped when this prints nothing, its
merge base from this run's working notes:

```bash
git -C <worktree> diff --name-only <merge-base>..HEAD ':(exclude)spectre/changes/<name>' ':(exclude)spectre/changes/archive/<name>'
```

The `## Report` names a skipped worktree as `- <worktree> — skipped, nothing outside its change
directory`.

Resolve the commands `project-get.sh <worktree> lint` and `project-get.sh <worktree> test` print
(auto-detect on exit 1).
Export the `KEY=value` lines `prepare-workspace.sh` printed for that worktree, then run the lint
commands, then the test commands, in the order printed, then `check-spec-reach.sh <worktree>` —
one more command in the same list, whose exit 0 line `Spec reach: not configured` is the ordinary
case for a project with no `regression checkout` (its header is canonical for its exit codes). Run
every command in order and do not stop at the first failure.
The lint commands with `check-spec-reach.sh` and the test commands run as two groups at once, each
in the order printed, each started detached and polled to completion as step 7 of the verifier's
steps (`skills/flow/visual-verify-verifier.md`) runs a `verify` that can outrun the shell tool's
time limit; the `## Report` lists every command in the order printed. **Nothing runs them later** —
`/flow`'s integrate phase has no verification gate — so a non-zero exit blocks this handoff.
While a stack is up for this run, the test commands start only after `check-dev-stack-fresh.sh
<worktree>` exits 0 or 2; exit 1 restarts the stack first, by the start rule of **Resolve
the run instructions** below, so no suite runs against a build older than the last commit; a stack
that will not start ends the turn as a failed command's environment cause does.

```text verified:design.md section 2 of this change
## Report
- `<command>` — exit <n>
  <the command's output, verbatim, or its last 20 lines when longer, stated as truncated>
```

The parent writes this `## Report` itself and shows it as this stage's output.
`prepare-workspace.sh`, both `project-get.sh` calls and this stage's `begin` mark are one Bash
call; the lint and test run, this run's `flow record dispatch begin`, and the ledger render below
are one more.

`run-verify-list.sh <worktree> lint` and `run-verify-list.sh <worktree> test` are the two groups: each runs its list as above, stdin detached, and prints that list's `## Report` lines, its exit codes canonical in its header — never a runner built by hand. On a change whose diff touches `ui paths`, the `test` group also ends with the project's `## visual verification` `verify` row, the full suite, run from the change's worktree of the `regression checkout` repo, unless decision.json records the visual verify skipped.
The test group skips what the list's last green run on a clean tree already covered: every command
when HEAD is the sha that run recorded, otherwise each command whose `# inputs:` pathspecs the delta
since that sha leaves untouched; a skipped command's `## Report` line says so, and the script's
header is canonical for the rule.

**Inline verify — a failing command.** A non-zero exit from any command in the list earns **one**
inline re-run of that command — the environmental-flake case. Output carrying `WARNING: DATA RACE`
or `race detected during execution of test` is never that case: the command takes **The loop**
whatever its re-run exits, since a race that passes once is still a race
(`rules/fix-determinism-at-the-source.mdc`), and `run-verify-list.sh` marks its block with
`data race — never environmental`. A second non-zero exit from the same
command is the branch's own defect and takes **The loop** (`skills/flow/verify-fix-loop.md`) once
the whole list has run, its first run's `result files:` and `full output:` paths named in the
fixer's dispatch; where the environment itself failed — a missing build prerequisite — it
ends the turn with `## Question` naming the command and its output, verbatim. A failing test this
change did not introduce takes **The loop** as well (**the sweep**, `skills/flow-contracts/known-bugs.md`). Never treat a
passing re-run as license to skip the rest of the list — every remaining command in the order
above still runs.

**Recording.** One `dispatches` row per worktree the list runs for, `-role verifier -key verify -model <parent
model> -effort <parent effort> -agent-id inline`, suffixed `-<repository name>` when this
run's resolved set holds more than one worktree — the same convention
**The verifier dispatch** (`skills/flow/visual-verify.md`) uses for the verifier's rows.
`run-verify-list.sh` records the row itself, never the parent by hand: both groups are run with the
same five trailing arguments, and each group records `begin` before its first command and ends the row
itself — `-outcome completed` when its list exited 0, `-outcome skipped` when every command of a test
group was cache-skipped, `-outcome blocked -cause test-failure` when its list exited non-zero. A
`begin` the store refuses stops the list before its first command, exit 4, its stderr line naming the
refusal: clear the cause — a tracked modification, most often — and run the list again; a failed
`end` is a warning line that never changes the exit code. The first `begin` stamps the start and the
later group's `end` the finish. A verify re-run after a new code commit passes a key no earlier verify row of the change
used — `verify-<k>`, `<k>` counting from 2, suffixed as above — because the store keeps a reused
key's first `startedAt`, and a row started before the commit never covers it (kan-1037). Once the one inline re-run is decided, a passing re-run ends a fresh
row `-outcome completed`, and a command that failed twice for a missing build prerequisite is ended
`-outcome blocked -cause environment` by the parent.

```bash
run-verify-list.sh <worktree> lint <name> verify[-<worktree basename>] mf-<literal-token> <parent model> <parent effort> [<canonical worktree>]
run-verify-list.sh <worktree> test <name> verify[-<worktree basename>] mf-<literal-token> <parent model> <parent effort> [<canonical worktree>]
# … only after a group exited non-zero and its re-run is decided:
flow record dispatch end -C <canonical worktree> -change <name> -key verify[-<worktree basename>] \
  -session-token mf-<literal-token> -outcome completed|blocked [-cause <cause>]
```

`<canonical worktree>` is the worktree holding `<project>/spectre/changes/<name>`; pass it when verifying any other worktree of a multi-repo change.

### Live check

**Runs only when `design.md` carries a `## Live check` section naming something to exercise**
(**D. Basic Workflow #3 — Writing plans**, `skills/flow/brainstorm-planner.md`); a one-line
no-runtime section, or none, skips it. A section is that no-runtime form when its first line opens
with `None`, `No ` or `Not applicable`, however many lines its justification wraps to. Once both groups of the lint and test list have exited —
never while either still runs: kan-918's live check, run beside the polled groups, wrote
`live-verification.md` into the tree `run-guard-tests.sh` watches, failing its watched-tree check
and costing a 272-second re-run — with the parent's own Bash calls: when the stack's worktree-resolved URLs do not answer, or `check-dev-stack-fresh.sh
<worktree>` exits non-zero, start it as **One stack for the verification tasks**
(`skills/flow/implement.md`) does; then exercise what the section names and write each
before/after figure, beside the failure look the section states, to `live-verification.md` in
the change's directory (**A change's directory**, `skills/flow-contracts/pipeline.md`), a
planning path the write-in-progress planning commit carries. **A before figure is measured from a
detached worktree at the merge base** — `git -C <worktree> worktree add --detach
<project>-worktrees/<name>-before <merge-base>`, the base's stack run from there, then stopped, and the worktree
removed with `git worktree remove` once the figure is written — never by checking base sources out
over the change worktree. A before figure already in
`<abs-worktree>/.superpowers/sdd/live-before.md`, measured while pass 1's panel read
(`skills/flow/review-panel.md`), is copied into the record only when the merge-base sha it records equals the current `git -C <abs-worktree> merge-base` — a round-boundary rebase moves the base (`skills/flow/review-panel-fix-round.md`) — and is measured again otherwise. Then run
`check-live-verification.sh
<changeRoot>/live-verification.md`; its header is canonical for the record it accepts. Exit 1 → the
record carries no before figure, or states one was not measured: measure it from the merge base as
above and rewrite the record, never proceed past it (kan-998). The check is never deferred past this stage:
`integrate-change.sh prepare` stops at `STOP: live-check` while the section names something to
exercise and `live-verification.md` is missing or fails `check-live-verification.sh` (kan-1019).
Exit 2 → it cannot answer; correct
the input and run it again. The `## Report` gains one line, `live check — <matched | failure look>`. A figure that
shows the failure look is the branch's own defect and takes **The loop**
(`skills/flow/verify-fix-loop.md`); a stack that will not start ends the turn as a failed
command's environment cause does. **A step of the check the operator waives runs on its cheapest safe proxy instead** — a scratch
database, a throwaway container — never simply dropped, and the `## Report` line names the proxy:
kan-576's waived V34 DOWN-block step, run by a reviewer on a scratch database, caught a rollback
that would have kept every `WORKOUTS: READ` grant as a full one (KAN-740). The verifier row's `end` (**Recording.**, above) is recorded
after this check, not before it: `-outcome blocked -cause test-failure` on a failure look,
`-cause environment` on a stack that will not start. The stack stays up for **Visual
verification** and **Resolve the run instructions**.

### Close the stage

**Load `skills/flow-contracts/session-records.md`** before reading the render outcome below.

**Confirm this run recorded a ledger** — rendering into the canonical worktree, the member of
this run's resolved set whose own change directory (**A change's directory**,
`skills/flow-contracts/pipeline.md`) holds `tasks.md` (the same member **1. Check for unfinished work**, `skills/flow/integrate.md`, passes
check-unfinished-work), never into whichever worktree this pass runs in, so a multi-worktree run
writes one copy, not one per worktree:

```bash
flow record render -change <name> -kind ledger -repo <canonical-worktree>
```

Read the outcome word, not the exit code. `rendered: <dest>` is ordinary. **`MISSING: ledger — no
rows for <name>` means this run recorded no dispatch at all**, reported plainly here rather than
discovered later. The outcome words are the table under
**Rendering the session records** (`skills/flow-contracts/session-records.md`).

**The live check gates the stage end.** `flow stage end` runs `check-live-verification.sh`'s
check in-process on the change's directory under `-C`, and refuses a `completed` end on its exit 1
or 2; its header is canonical for the verdict. A refusal on exit 1 → run **Live check** above and
rewrite the record, never end the stage past it; on exit 2 → pass `-C` the canonical worktree, or
correct the input it names, and end again (kan-1008: the stage end checked nothing, the live check
first ran at integrate, and its defect landed after the panel; kan-1029: a stage end with no
chained check passed).

```bash
flow stage end -C <canonical-worktree> -command '/flow' -stage flow.verify -outcome completed <name>
```

## Visual verification

```bash
flow stage begin -command '/flow' \
  -stage flow.visual-verify \
  -harness <harness> \
  -session-token mf-<literal-token> \
  <name>
```

Reads the `## visual verification` section, canonical in
`skills/flow-contracts/project-configuration-visual.md`. This stage owns its procedure — nothing else in
this pipeline restates it. Resolve once per worktree in this run's resolved set, the same set
**Verify** above resolved:

1. **Resolve the section — with the guard, not by eye.** Run

   ```bash
   git -C <worktree> diff --name-only <merge-base>..HEAD | check-visual-trigger.sh <worktree>
   ```

   Exit 2 with a `VISUAL-TRIGGER-NOT-CONFIGURED:` line on stderr → this worktree is skipped for the rest of this stage.
2. **Match the diff** — the same call's verdict.
   Exit 0 → at least one changed path matched a declared `ui paths` glob; continue. Exit 1 → this
   worktree is skipped for the rest of this stage. Exit 2
   with a `VISUAL-TRIGGER-CANNOT-ANSWER:` line → report its stderr and skip this worktree the same
   way exit 1 does.
   `check-visual-trigger.sh` owns the glob semantics (`**` spanning directories, a leading
   dot-slash prefix, an absolute glob, a glob with a space); nothing here restates them.

   In a run whose resolved set holds more than one worktree, every other worktree's diff is also
   matched against the canonical worktree's `ui paths`, each path prefixed with that worktree's
   repository name — its main checkout's basename, the name a glob such as `gymie-frontend/**`
   spells another repository's files with:

   ```bash
   git -C <worktree> diff --name-only <merge-base>..HEAD | check-visual-trigger.sh --repo <repo-name> <canonical-worktree>
   ```

   Its exit 0 makes the canonical worktree survive this step as the canonical worktree's own exit 0
   does.

   `<merge-base>` in both calls is that worktree's own merge base from this run's working notes —
   never `origin/<base>` or another ref picked by hand, which matches every path the base gained
   since the worktree was cut as this change's own (KAN-702).

   **On a fix run, both calls read the fix run's own diff.** A worktree that has a since-close sha this run — the one **The late-fix reduction** (`skills/flow/review-panel-late-fix.md`) read for it — feeds them `git -C <worktree> diff --name-only <since-close-sha>..HEAD` in place of `<merge-base>..HEAD`, so a fix run whose delta touches no `ui paths` skips this stage at step 2 as exit 1 does.
   A worktree whose since-close sha is `-` keeps `<merge-base>..HEAD`, and so does every worktree when the change's newest verifier dispatch row (the newest row `flow record dispatches -change <name>` returns with `role` `verifier` and a `key` starting with `visual-verify` — the rows `<agents repo>/scripts/check-visual-verify-dispatched.sh` counts) does not have `outcome` `completed` — an earlier run's UI change that was never verified may lie before the since-close sha.

**A `skipped` decision ends the stage at step 2.** When this run's decision
(`<abs-worktree>/.superpowers/sdd/decision.json`) records `visual.verify` `skipped` (step 6 of
**Decide**, `skills/flow/brainstorm-planner.md`), every worktree surviving step 2 prints
`Visual: skipped at Decide — <reason>`, nothing is dispatched, and the stage closes:

```bash
flow stage end -command '/flow' -stage flow.visual-verify -outcome skipped <name>
```

`required`, `not configured` or no `visual` at all takes the load below, unchanged.

**Load `skills/flow/visual-verify.md`** once at least one worktree survives step 2, and run it
from step 3 for every surviving worktree — it carries **The verifier dispatch**, steps 3–13, this
stage's `## Report`, **Blocking**, and this stage's `end` mark. When no worktree survives step 2,
nothing is dispatched and that file is never loaded; close the stage here:

```bash
flow stage end -command '/flow' -stage flow.visual-verify -outcome completed <name>
```

## Stage, excluding the planning paths

```bash
flow stage begin -command '/flow' -stage flow.stage-diff -harness <harness> -session-token mf-<literal-token> <name>
```

Confirm every intended task checkbox is `[x]`, and that `git log <merge-base>..HEAD` shows one
commit per completed task plus each fix round's fix commits on top — a pushed branch takes a
panel fix as a new commit, never a rewrite (**Panel re-runs**, `skills/flow/review-panel-fix-round.md`) —
with every red-task-partner and unpushed-history fixup already folded in via
`git rebase --autosquash`; no stray `fixup!` commit should remain unsquashed, unless a PR already
exists (below).

In **every** affected worktree:

```bash
git -C <worktree> status
git -C <worktree> log <merge-base>..HEAD --oneline
```

> **`<project>/spectre/changes/` is never part of a task commit.** `<project>/spectre/specs/`
> is not planning — a capability spec belongs in the task commit that implements its
> requirement. This step only confirms nothing slipped in.

**Load `skills/flow-contracts/git-boundaries.md`** before committing below.

**The PR-branch split.** Every task, fixup and planning commit already sits on the branch,
pushed as it landed (**Branch backup**, `skills/flow-contracts/git-boundaries.md`). If the state
file records a `prUrl`, a PR is already open, so this run also commits whatever the operator
edited at the human gate and whatever planning delta the last planning commit left, and pushes
everything to the PR branch; otherwise this step commits nothing more.

**Load `skills/flow-contracts/git-boundaries-commit-chain.md`** only when the state file records a
`prUrl` — `commit-split.sh` below runs that chain.

On that path only — and in this order — run
`flow record render -change <name> -kind all -repo <canonical-worktree>` (the same member the
ledger render above targets); then `commit-split.sh <worktree>
<name> "<impl-msg>" "chore(spectre): plan <name>"`; then push the branch
`--force-with-lease`, since the split reshaped it. `<impl-msg>`
covers working-tree edits the operator made at the human gate without staging them — derive it the
same way a fixup commit's subject is derived — `fix(<module>): <what changed since the last task
commit>`.

The render overwrites in place. **A non-zero exit means a destination was refused or could not be
written** — report it, and continue committing the fix.

```bash
flow stage end -command '/flow' -stage flow.stage-diff -outcome completed <name>
```

## Resolve the run instructions

```bash
flow stage begin -command '/flow' -stage flow.run-instructions -harness <harness> -session-token mf-<literal-token> <name>
```

Resolve the run instructions the start below uses, the handoff's `**Running:**` section prints (**Write `IN_PROGRESS`**, below) and `/flow-status` prints as its `Running:` section. It writes no file.

- **Every start command comes from `<project>/.flow/project.md`'s `## run`**, with every path
  made absolute, each started as `<agents repo>/scripts/run-capped.sh -- <command>`.
- **Every URL is the one this worktree resolved**, never the project's declared base. Resolve each
  URL from this worktree's workspace id. A project that declares no isolation resolves nothing.
  An application whose port is fixed outside that project's own repository keeps its default,
  named with a short note.
- Apps in scope come from `## apps` in `<project>/.flow/project.md`, or auto-detection.
- **Where the project declares no runnable application**, resolve the `## lint` and `## test`
  commands instead.
- **Before this stage ends, start the stack.** This runs on every run — first and fix alike, not
  only a fix run. A fix run's motivation is the sharpest example: it hands the operator a diff and
  the run instructions this stage resolves, and if the applications those instructions name are
  still serving pre-fix code, the operator reviews one thing and runs another. But the same
  applies on a first run: a handoff is meant to hand off a running stack, not a command that starts
  one, so before this stage ends, rebuild and restart every application the run instructions above
  name — the ones just resolved, and no others — from the project's `## run` commands. **A stack already serving this
  worktree's build is left running**: when its URLs answer and `check-dev-stack-fresh.sh
  <worktree>` exits 0 before the start, the start is skipped — the stack the verification tasks
  or **Live check** started is the one handed off. Exit 1 or 2 starts it as above.

  **A run that will land straight away starts nothing.** When the record's `reviewBeforeLanding`
  is not `true` and the run leaves no `?` line for its `Decisions:` (**Write `IN_PROGRESS`**,
  below), no handoff hands a stack to anyone: the start, the stack freshness check and a refused
  start's stop below are all skipped, so a held port never keeps the change from landing (kan-1030).

  Resolve the start from the two keys the project already declares, in order: its `## stop`, when
  it declares a command, then its `## run`. **A `## stop` that deliberately declares no command
  means there is nothing to stop, not that this rule is skipped** — the start is then whatever
  `## run` does to bring the application up fresh. No new project-configuration key is added for
  this, and none is needed.

  **Never the flow dev stack.** `flowd` on `127.0.0.1:4173`, its `flow-postgres` container and the
  `flow` database inside it are never stopped, restarted or dropped by any run —
  `<project>/CLAUDE.md` states that prohibition and this rule does not weaken it. Where a project's
  own `## run` names that service, the prohibition wins over this start rule, never the reverse.
  This is separate from the visual-verification procedure's own start/stop rule (step 13, `skills/flow/visual-verify.md`): that stage
  stops only the stack it started for its own probe, and that rule is not restated here. This rule
  starts whatever the run instructions name, on every run, regardless of whether that stage ran
  or started anything.

  **Where every application `## apps` names is one this prohibition covers, the start is
  nothing — stated, not silently skipped.** The handoff then names the application skipped and
  why:

  ```
  Not started: <app> (<url>) — protected, see <project>/CLAUDE.md.
  ```

  **A start that fails blocks this stage**, naming the application and what the command printed —
  handing over run instructions that cannot be followed is the failure this rule exists to prevent.
  Where the project declares no runnable application, there is nothing to start and the rule is
  satisfied by saying so, not by silently skipping it.

- **The stack behind the URLs is checked, not trusted.** When the run instructions carry URLs, run
  `check-dev-stack-fresh.sh <worktree>` first: the project's declared
  `fingerprint` row (the `## visual verification` section) proves what the stack serves is the
  worktree's own build. The guard's own header is canonical for what its three exits
  cover. Exit 0 adds nothing. Exit 1 adds one `Summary:` bullet, naming the application, its
  resolved URL and the guard's own stderr reason, restart-shaped and never a restatement of the
  verdict: `Stale: <app> (<url>) — <the guard's reason>; restart the stack before testing.`
  Exit 2 adds `Freshness: unverified — <the guard's stderr reason>` instead, the same
  visible-gap rule the visual-verification procedure's own step 6 (`skills/flow/visual-verify.md`) runs on a missing row. A refused start
  (below) already ends the run, so this check only ever runs on a start that succeeded or was
  skipped by the protected-service rule, or on the rest of a stack the out-of-worktree exception
  below hands off.

- **A refused start is relayed, never resolved.** When the start command above exits non-zero
  because a port it needs is already held, this stage does not retry, does not pick another port,
  and does not stop the holder. It prints that command's output verbatim, then the holder's own
  `devStop`-style stop command when the output names the holder's worktree (or, when it does not,
  a note to run `check-worktree-processes.sh`-style diagnosis and stop the holder's stack by hand).
  This run then ends at `IN_PROGRESS` with the diff still staged, and the handoff names `/flow
  <name>` as the next command — the same change, re-run once the holder is stopped. It never stops
  a stack this run did not start.

  **An application outside the change's worktrees is the one exception.** When the refused start
  is that of an application whose repository the change holds no worktree in — its code is not
  under review — the rest of the stack is handed off, and the `Running:` section names it in place
  of a URL bullet:

  ```
  Not started: <app> (<url>) — port <port> held by <the holder the start's output names>; outside this change's worktrees.
  ```

```bash
flow stage end -command '/flow' -stage flow.run-instructions -outcome completed <name>
```

## Write `IN_PROGRESS`

```bash
flow stage begin -command '/flow' -stage flow.write-in-progress -harness <harness> -session-token mf-<literal-token> <name>
```

**Append this run's own narrative first.** Append to `narrative.md` in the change's directory
under `<abs-worktree>` (**A change's directory**, `skills/flow-contracts/pipeline.md`) (create it
with the title `# <name> —
session narrative` when absent) one section `## <YYYY-MM-DD> — <creating run | fix run>` holding
this session's own prose account of the run — problems hit, workarounds, time sinks, environment
gaps, operator decisions taken mid-run, every approach tried and abandoned with why it failed —
and nothing the ledger or panel record already holds. The abandoned approaches and their failure
reasons are recorded because the diff shows only the winning shape — a deferred self-review pass
reads the narrative, and the negative results live nowhere else in the change's record. A
durable process lesson this run paid for is promoted to the lessons home (**Process lessons**,
`skills/flow-contracts/lessons.md`) — the brief is repo content, landed with this change's own
work. The write-in-progress planning commit carries it (**Planning commits**,
`skills/flow-contracts/git-boundaries.md`), made once the append lands; nothing else stages it.

**Then check the narrative against the runs the store marked:** `check-narrative-runs.sh <abs-worktree> <name> <narrative.md>`. Exit 1 names a fix or integrate run whose `## <YYYY-MM-DD> — fix run` or `— integrate run` section the file lacks: append that run's section from the change's ledger and history, and run the check again. Exit 2 is reported and stops the write.

**Close every open dispatch row of the change** before the state is written — integrate's
**1. Check for unfinished work** (`skills/flow/integrate.md`) stops a landing on one, and only this
run knows how its own dispatches ended. List them from the canonical worktree:

```bash
flow record dispatches -change <name> -C <canonical-worktree> \
  | jq -r '.[] | select((.outcome // "") == "") | "\(.key) \(.sessionToken)"'
```

and close each printed row under its own key and token, `-outcome completed` when its report was
read, `-outcome aborted` otherwise — an earlier run's row included:

```bash
flow record dispatch end -change <name> -C <canonical-worktree> -key <key> -session-token <its token> -outcome <completed|aborted> \
  -ended-at <when its work ended> [-commit <its task commit>]
```

`-ended-at` is the instant the row's work ended — its task commit's committer date
(`git log -1 --format=%cI <sha>`), else its report file's mtime — never the moment it is closed,
and an implementer row carries `-commit` with its task commit's full sha.

**Record the `**Summary:**` bullets next**, per **Summary and live-stack line, before every handoff** (`skills/flow-contracts/pipeline.md`), since the guard below finds a change with no recorded summary unfinished.

**Every run writing `IN_PROGRESS` then runs the unfinished-work guard on itself**, a creating run
as well as a fix run, once per worktree of the resolved
set, with the arguments **Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`) passes it —
`check-unfinished-work.sh <worktree> <name> <canonical-worktree> <worktree>=<merge-base>...` — so an appended task left
unrecorded or unreviewed, or a panel round dispatched with no pass row, stops this run rather than the next integrate. An `OUTSTANDING:` line is
this run's own unfinished work: finish what its breakdown names — a missing dispatch or review is
run and recorded, never recorded after the fact for work that did not run — and run the guard
again, per **Finish it — never a prompt** (`skills/flow/unfinished-work-gate.md`), whose false-positive record covers an item no change can finish; the state is written once every worktree prints `CLEAR:` or every remaining item is so recorded — never a stop or a question. A missing verdict line or an
exit 2 is reported and stops the run at the state it read.
The guard folds in `check-visual-verify-dispatched.sh`'s verdict for its own worktree and finds a
code commit no later completed `verify` row covers unfinished, so a skipped visual verification or
an unrecorded inline verify stops this run too.

**Every run then runs the visual guard on itself** — creating, resuming and fix runs alike — once
per worktree of the resolved set, with the arguments integrate's gate passes it (**Run 1 — the
branch is not merged**, `skills/flow-contracts/finish-contract-run1.md`):

```bash
check-visual-verify-dispatched.sh <worktree> <name> $(git -C <worktree> merge-base HEAD origin/<base>) <canonical-worktree>
```

A `VISUAL-VERIFY-MISSING:` line means a UI-path commit landed after the last HEAD a verifier saw —
a visual-fix task's, a panel fix's, a full-suite fix's or a fix run's own. Run **Visual
verification** above again from its `begin` mark on the current HEAD, take every departure through
**The loop** (`skills/flow/verify-fix-loop.md`), and run the guard again; the state is written,
and the handoff printed, only once every worktree prints `VISUAL-VERIFY-OK:`. When the line names a
verified HEAD and the paths changed since it, that re-run is scoped as a round of **The loop** is
(`skills/flow/verify-fix-loop.md`): the named HEAD is its round base, its `motions:` names only the
motions `git diff <that HEAD>..HEAD` adds, changes or removes, and a clean scoped report closes it,
since that HEAD's full run stands. Integrate stops on
the same verdict and no operator course there passes it. An exit 2 is reported and stops the run
at the state it read.

Write the state file: `IN_PROGRESS` from `STARTED`, otherwise **the state exactly as read**.
`worktrees` should already carry one absolute-path key per affected worktree and its merge base,
so this step re-reads the current record and
confirms every resolved worktree is present rather than reconstructing the map from scratch;
add any entry still missing (a worktree added after the last incremental write) before
proceeding. Carry `artifactUrl` (always `null` under `/flow`), `jiraIssue`, `planningEffort`
(never written — `null` unless a legacy record holds a level), `models.default` (a legacy field,
likewise never written — `/flow` resolves models from the settings
store per run, never records a value into the per-change state), `reviewBeforeLanding` and `prUrl`
forward verbatim.
The state file lives outside the repo — never `git add` it.

```bash
flow stage end -command '/flow' -stage flow.write-in-progress -outcome completed <name>
```

**Then land straight away unless review was asked for or a question is open** — a run that lands re-records first any
`**Summary:**` bullet changed after the guard, since no printed block records it again. Re-read the record:

```bash
flow state get <name> | jq -r '.reviewBeforeLanding // false'
```

`true` prints the block below and ends the run. **So does a run that leaves a `?` line for its
`Decisions:`, whatever the record holds** — an open question the operator must answer is never
landed past (kan-1030). Anything else prints no block but the block's `**Running:**` section,
always, and its 🤖 `Decisions:` lines, when it took any, under a `**Decisions:**` heading —
integrate's handoffs carry no such field — and continues at **Deciding which run this is**
(`skills/flow/integrate.md`) in the same turn, as if a bare `/flow <name>` had started it; that
run prints its own handoff or stop. This is the one chain into integrate **Every invocation is
re-entrant** (`skills/flow-contracts/pipeline.md`) sanctions.

````
**Summary:**
- <what this round changed for a user of the app, in product terms — a few bullets at most>
- <a line another contract requires in the handoff — a caveat, a manual check, a push command, an open defect — one bullet each>

**Running:**
- <app>: <the URL this worktree resolved> — one bullet per started application
- Login: <app> — <user> / <password> — one bullet per `## credentials` row, when the section declares any
- Not started: <app> (<url>) — protected, see <project>/CLAUDE.md. | Nothing to start — no runnable application; <the `## lint` and `## test` commands>
- Not started: <app> (<url>) — port <port> held by <holder>; outside this change's worktrees. — one bullet per application **A refused start is relayed, never resolved** above hands off unstarted

**Decisions:** none | <one line each:>
- ? <an open question the operator must answer>
- 🤖 <question> → <the recommended option taken>

Next:
/clear
```
/flow <name>
```
````

**This block is the whole handoff of a creating, resuming or fix run that stops at `IN_PROGRESS`**
— no heading, no other field, and no summary or live-stack line before it (**Summary and
live-stack line, before every handoff**, `skills/flow-contracts/pipeline.md`); bullets that changed after the guard — a visual-fix task, a re-verify — are recorded again as the block prints. A stop — a failure,
a blocked step, a refused start, a `## Question` — prints as its own rule states, not this block.

**The `Running:` section is never omitted** — not from this block, and not from the chain into
integrate above. It carries a URL for every application the run instructions name, the admin panel
included, and a login for every account the operator needs to reach them. Its URLs are the ones **Resolve the run
instructions** above resolved for the stack it started or left running, never the project's
declared default port; its logins are every row `<project>/.flow/project.md`'s `## credentials`
declares. Where nothing was started, its last line states why in place of the other two.

**The `Decisions:` lines name every prompt this run answered itself** on its recommended
option (**Auto-resolution**, `skills/flow-contracts/operator-prompts.md`), so the operator can
overrule any of them with a fix run; it reads `none` when the run took none and left no open
question. Planning answers the `## decisions: recommended` mode took are named here the same way.

**No `**Learning:**` or `**Prompt tuning:**` line** — both are shown only once the change is fully
done, in the Finished handoff (`skills/flow/cleanup.md`).

**The parent prints this block directly, as this stage's own output** — no return, no relay: the
running session assembles it here and shows it to the operator in the same turn (**The parent
orchestrates directly**, `skills/flow/implement.md`).

## Guardrails

- **Never** create a second worktree for the same change.
