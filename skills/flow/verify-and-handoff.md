# Verify, stage, and hand off

Loaded by `skills/flow/SKILL.md` once `skills/flow/review-panel.md` closes clean. Stage order:
**verify → visual-verify → stage-diff → write-in-progress**.

## Verify

**Load `skills/flow-contracts/worktree-resolution.md`** before resolving this run's worktree set,
below.

```bash
flow stage begin -command '/flow' \
  -stage flow.verify \
  -harness <harness> -session-token mf-<literal-token> \
  <name>
```

**First, validate the section and export what it declares — with the script, not by eye.** Run

```bash
flow-guard prepare-workspace <worktree>
```

once per worktree in this run's resolved set — the same set **2. Isolate the workspace**
(`skills/flow/implement.md`) resolved, non-empty by construction — never a raw read of the state
file's `worktrees` map. Per **Resolving a change's worktrees**
(`skills/flow-contracts/worktree-resolution.md`), report an empty resolved set and do not proceed.

`flow-guard prepare-workspace` validates the worktree's isolation table first, then — only
if that passes — derives and exports the variables the project's `## workspace isolation` section
declares, resolved against the workspace id, and prints one `KEY=value` line per exported variable
to stdout. Exit 0 means the printed lines are what to carry forward into `## lint` and `## test`
below (nothing printed means the project declares no `## workspace isolation` section). A non-zero
exit is the dropped-row case (exit 1, relay the script's own lines verbatim and stop) or the
cannot-answer case (exit 2, stop the same way) — stop **before** `## lint` and `## test`, without
writing the state file.

**Load `skills/flow-contracts/workspace-isolation.md` by section, only the one each trigger
needs** — `grep -n` for its heading plus `sed -n` for that section, never the whole file: **The
cache index** when `flow-guard prepare-workspace` exited 0 with stderr naming a `cache index` row, **The
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
flow-guard check-base-behind <worktree>...
```

over the same resolved set; its header is canonical for its lines. Exit 0 → continue. Exit 1 → run `flow-guard sync-base-behind <worktree>...` — the guard itself merges the tracked branch into the base, pushes the base (unasked: **What still stops**, `skills/flow-contracts/operator-prompts-auto-resolution.md`; the guard refuses an empty merge commit, whose refspec would delete the remote base, kan-1087 self-review F4), fetches and syncs the worktree onto it; never a raw `git push` — then run **Verify** again from its `begin` mark. Exit 2 → name its stderr in the handoff and continue.

**This step does not call the project's `create` command.** `create` is called by whatever starts
the project's applications, per **Project configuration**
(`skills/flow-contracts/project-configuration.md`), and this step starts them only for **Live check** below, through the project's own `## run`.

**After the panel closes, the parent edits no source outside the fix procedure.** Source
changes only through the fix procedure (`skills/flow/fix.md`) or a fix run the operator
starts; any other source change from here on makes every slot's result stale (**Panel re-runs**,
`skills/flow/review-panel.md`).

**Load `skills/flow/fix.md`** only when this stage's or `flow.verify`'s final
report carries a fixable defect — it carries the whole procedure.

### Inline verify

**A worktree with nothing outside its change directory runs none of the list below.** In a run
whose resolved set holds more than one worktree, a worktree is skipped when this prints nothing, its
merge base from this run's working notes:

```bash
git -C <worktree> diff --name-only <merge-base>..HEAD ':(exclude)spectre/changes/<name>'
```

The `## Report` names a skipped worktree as `- <worktree> — skipped, nothing outside its change
directory`.

Resolve the commands `flow-guard project-get <worktree> lint` and `flow-guard project-get <worktree> test` print
(auto-detect on exit 1).
Export the `KEY=value` lines `flow-guard prepare-workspace` printed for that worktree, then run the lint
commands, then the test commands, in the order printed — `flow-guard run-verify-list` appends
the spec-reach check to the lint list itself (`Spec reach: not
configured` is the ordinary case for a project with no `regression checkout`). Run
every command in order and do not stop at the first failure.
The lint commands and the test commands run as two groups at once, each
in the order printed, in the one Bash call of **Recording** below, which a runner
(its row of **Dispatch sites — the parent's closed list**, `skills/flow/implement.md`; `runner-<n>`, the call as its
`COMMANDS:`) runs and answers with its verdict — the parent writes the pair it dispatches the runner on into those arguments as `<runner model>` and `<runner effort>`, so the `verify` rows show the runner's pair; the `## Report` lists every command in the order
printed. **Nothing runs them later** —
`/flow`'s integrate phase has no verification gate — so a non-zero exit blocks this handoff.
While a stack is up for this run, the test commands start only after `flow-guard check-dev-stack-fresh
<worktree>` exits 0 or 2; exit 1 restarts the stack first, from the project's `## stop` when it declares a command, then its
`## run`, so no suite runs against a build older than the last commit; a stack
that will not start ends the turn as a failed command's environment cause does.

**A `verify` that can outrun the shell tool's time limit runs whole, in the runner's foreground.** Never split its
`<specs>` into hand-chosen chunks to fit the limit: a hand-made split can drop a spec, and each
chunk pays the suite's start-up again.

```text
## Report
- `<command>` — exit <n>
  <the command's output, verbatim, or its last 20 lines when longer, stated as truncated>
```

The parent writes this `## Report` itself and shows it as this stage's output.
`flow-guard prepare-workspace`, both `flow-guard project-get` calls and this stage's `begin` mark are one Bash
call; the lint and test run, this run's `flow record dispatch begin`, and the ledger render below
are one more.

`flow-guard run-verify-list <worktree> lint` and `flow-guard run-verify-list <worktree> test` are the two groups: each runs its list as above, stdin detached, and prints that list's `## Report` lines, its exit codes canonical in its header — never a runner built by hand. On a change whose diff touches `ui paths`, the `test` group also ends with the project's `## visual verification` `verify` row, on the specs its `specs` row maps (the smoke set alone, `--grep @smoke`, with no `specs` row or one that maps nothing; the whole suite only when the operator asks), run from the change's worktree of the `regression checkout` repo, unless decision.json records the visual verify skipped.
The test group skips what the list's last green run on a clean tree already covered: every command
when HEAD is the sha that run recorded, otherwise each command whose `# inputs:` pathspecs the delta
since that sha leaves untouched; a skipped command's `## Report` line says so, and the script's
header is canonical for the rule. At the same HEAD it also skips each command an earlier run on a clean tree passed while another of that run's commands failed, its `## Report` line naming that sha, so a re-run after one failure runs only what has not passed.

**The full suite is `flow-guard run-verify-list -full <worktree> test`, run only when the operator
asks** (**Full-suite instruction**, `skills/flow/SKILL.md`). It is the same list with every skip
off: every suite runs, no command is skipped on a green run, and the visual `verify` row runs whole. Its suite rows are recorded as
`<name>-full`, and with the recording arguments its row stands as the `verify` gate. A command
it fails that the default list passed at the same HEAD is marked `escaped the default list`; the
script's header is canonical for the rest.

**Only the operator starts the full suite.** `-full` is refused unless `FLOW_OPERATOR_FULL_SUITE=1`
is in its environment. The operator sets that variable; a session or a subagent never sets it, whoever
asks. A session asked for the full suite whose environment
lacks it hands the operator the command — `FLOW_OPERATOR_FULL_SUITE=1 <agents repo>/scripts/flow-guard
run-verify-list -full <worktree> test` — and goes on with the default list. Every `/flow`
and `/flow-interactive` run, in every project, runs the smoke set and the change's
targeted tests and nothing wider: a targeting that is empty or fails falls back to the smoke set,
never to a whole suite, and no whole-suite command is typed by hand in its place.

**Inline verify — a failing command.** A non-zero exit from any command in the list earns **one**
re-run by the runner, which reports `flaky`, `defect` or `environment` evidence
(`skills/flow/runner-brief.md`) — the environmental-flake case. A failing visual `verify` row's one re-run is the narrowed run of **Narrowed visual re-runs**
(`skills/flow/fix-visual.md`). Output carrying `WARNING: DATA RACE`
or `race detected during execution of test` is never that case: the command takes the fix procedure
whatever its re-run exits, since a race that passes once is still a race
(`rules/fix-determinism-at-the-source.mdc`), and `flow-guard run-verify-list` marks its block with
`data race — never environmental`. A failure the run traced to nondeterminism in code or a test — an unwaited kill, a read racing an exit, a zombie read as alive — is never that case either: it takes the fix procedure whatever its re-run exits, or, for a test of the flow tooling, an in-run pipeline fix (**Pipeline defects found mid-run**, `skills/flow-contracts/pipeline.md`). `flow-guard integrate-change commit` refuses a narrative line that leaves a failing or flaky test open. A second non-zero exit from the same
command is the branch's own defect and takes the fix procedure (`skills/flow/fix.md`) once
the whole list has run, its first run's `result files:` and `full output:` paths named in the
fixer's dispatch; where the environment itself failed — a missing build prerequisite — it
ends the turn with `## Question` naming the command and its output, verbatim. A failing test this
change did not introduce takes the fix procedure as well (**the sweep**, `skills/flow-contracts/known-bugs.md`). Never treat a
passing re-run as license to skip the rest of the list — every remaining command in the order
above still runs.

**Inline verify — the confirm run.** After a fix round, only the failing test commands re-run until
they pass (`skills/flow/fix.md`, step 3); then the default list runs once in full, and only that
run's green lets the change land. A command that fails in the confirm run opens the next fix.

**A test runner's own timeout under outside load is the environment's failure.** A command killed by
its test runner's own timeout — Go's `panic: test timed out after` — with no failed assertion and no
data race, while the one-minute load average (`uptime`) exceeds the machine's core count
(`getconf _NPROCESSORS_ONLN`), takes neither the fix procedure nor a `## Question`: it is re-run, and `flow-guard run-verify-list` holds
every `test` or `visual` list until that load average is at or under the core count (THE LOAD GATE,
its header canonical) — its exit 5, the load never having dropped, is re-run the same way. A
timeout on a run started at or under the core count is the branch's own defect and takes the fix procedure. No
timeout is raised to absorb the load (`rules/fix-determinism-at-the-source.mdc`).

**Every run of the list is the runner's — the first, the one re-run, a re-run after a fix.** The parent never types `flow-guard run-verify-list` for it itself: a run it types bills the main session's model while its `verify` row records the runner's pair, and the dashboard reads that row as a runner that never ran.

**Recording.** One `dispatches` row per worktree the list runs for, `-role runner -key verify -model <runner
model> -effort <runner effort> -agent-id inline`, suffixed `-<repository name>` when this
run's resolved set holds more than one worktree — the same convention
**The verifier dispatch** (`skills/flow/visual-verify.md`) uses for the verifier's rows.
`flow-guard run-verify-list` records the row itself, never the parent by hand: both groups are run with the
same five trailing arguments, and each group records `begin` before its first command and ends the row
itself — `-outcome completed` when its list exited 0, `-outcome skipped` when every command of a test
group was cache-skipped, `-outcome blocked -cause test-failure` when its list exited non-zero. A
`begin` the store refuses stops the list before its first command, exit 4, its stderr line naming the
refusal: clear the cause — a tracked modification, most often — and run the list again; a failed
`end` is a warning line that never changes the exit code. The first `begin` stamps the start and the
last group to finish ends the row: a group that ends green while a sibling group of the same key still runs records no `end`, and a group killed before its end leaves the row open for the parent to end by hand. The next verify run of the same group — a key that folds to the same `verify[-<repository name>]` — ends such a row `-outcome aborted` at its own begin. A verify re-run after a new code commit passes a key no earlier verify row of the change
used — `verify-<k>`, `<k>` counting from 2, suffixed as above — because the store keeps a reused
key's first `startedAt`, and a row started before the commit never covers it. `verify-<k>` keys a re-run with no fix; a re-run after a fix carries the fix round's `-fix-<k>` key
(`skills/flow/fix.md`). Once the one inline re-run is decided, a passing re-run ends a fresh
row `-outcome completed`, and a command that failed twice for a missing build prerequisite is ended
`-outcome blocked -cause environment` by the parent.

```bash
flow-guard run-verify-list <worktree> lint <name> verify[-<repository name>] mf-<literal-token> <runner model> <runner effort> [<canonical worktree>] & l=$!
flow-guard run-verify-list <worktree> test <name> verify[-<repository name>] mf-<literal-token> <runner model> <runner effort> [<canonical worktree>]; t=$?
wait "$l"; echo "lint: $? test: $t"
# … only after a group exited non-zero and its re-run is decided:
flow record dispatch end -C <canonical worktree> -change <name> -key verify[-<repository name>] \
  -session-token mf-<literal-token> -outcome completed|blocked [-cause <cause>]
```

Both groups go out in one Bash call under one key, as the block above writes them — never two calls in
turn, never a second key.

`<canonical worktree>` is the worktree holding `<project>/spectre/changes/<name>`; pass it when verifying any other worktree of a multi-repo change.

`flow-guard run-verify-list` runs `flow-guard prepare-workspace <worktree>` itself before its first command and gives every command the `KEY=value` lines it prints, a value the caller's environment already sets winning; the caller exports nothing.

### Live check

**Runs only when `design.md` carries a `## Live check` section naming something to exercise**
(**D. Basic Workflow #3 — Writing plans**, `skills/flow/brainstorm-planner.md`); a one-line
no-runtime section, or none, skips it. A section is that no-runtime form when its first line opens
with `None`, `No ` or `Not applicable`, however many lines its justification wraps to. Once both groups of the lint and test list have exited —
never while either still runs — with the parent's own Bash calls: when the stack's worktree-resolved URLs do not answer, start it as **One stack for the verification tasks**
(`skills/flow/implement.md`) does; then exercise what the section names — a proxy (an earlier capture, a neighbouring check) never stands for an item the section names unless that item is infeasible in this environment, and an unexercised item is exercised before the stage ends — and write each
before/after figure, beside the failure look the section states, to `live-verification.md` in
the change's directory (**A change's directory**, `skills/flow-contracts/pipeline.md`), a
gitignored planning path written on disk and nothing else. **A before figure is measured from a
detached worktree at the merge base** — `git -C <worktree> worktree add --detach
<project>-worktrees/<name>-before <merge-base>`, the base's stack run from there, then stopped, and the worktree
removed with `git worktree remove` once the figure is written — never by checking base sources out
over the change worktree. A before figure already in
`<abs-worktree>/.superpowers/sdd/live-before.md`, measured while pass 1's panel read
(`skills/flow/review-panel.md`), is copied into the record only when the merge-base sha it records equals the current `git -C <abs-worktree> merge-base` — a round-boundary rebase moves the base (`skills/flow/review-panel-fix-round.md`) — and is measured again otherwise. The `## Report` gains one line, `live check — <matched | failure look>`. A figure that
shows the failure look is the branch's own defect and takes the fix procedure
(`skills/flow/fix.md`); a stack that will not start ends the turn as a failed
command's environment cause does. **A step of the check the operator waives runs on its cheapest safe proxy instead** — a scratch
database, a throwaway container — never simply dropped, and the `## Report` line names the proxy. The verifier row's `end` (**Recording.**, above) is recorded
after this check, not before it: `-outcome blocked -cause test-failure` on a failure look,
`-cause environment` on a stack that will not start. The stack stays up for **Visual
verification**.

**The check runs at the smallest footprint that proves its claim** — the lowest parallelism, the
fewest inputs that still prove the claim — and widens only once that run
passes.

**A check command that outlasts the Bash tool's foreground limit runs with the tool's
`run_in_background`**, its log and exit status written under the change's directory (**A
change's directory**, `skills/flow-contracts/pipeline.md`) rather than a fixed `/tmp` path, and
the parent waits on the tool's completion notice — never a hand-built `nohup` and a polling loop.

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

## Visual verification

Reads the `## visual verification` section, canonical in
`skills/flow-contracts/project-configuration-visual.md`. This stage owns its procedure — nothing else in
this pipeline restates it. **`flow-guard visual-verify-pass prepare` runs steps 1
and 2 below (it makes no stage mark — `flow.verify` stays open — and begins each dispatch's row) in the parent's own Bash call** — never these steps by hand, and never a subagent's:

```bash
flow-guard visual-verify-pass prepare <canonical-worktree> <name> <key> --harness <harness> --session-token mf-<literal-token> \
  --merge-base <worktree>=<sha> [--worktree <worktree>] [--since <worktree>=<sha>] \
  [--round-base <worktree>=<sha>] [--defects <worktree>=<frame>[,<frame>…]] [--app <worktree>:<app-root>=<resolved-url>]
```

Its header (`<agents repo>/scripts/flow-guard visual-verify-pass`) is canonical for every argument, each step it
runs and its closing lines. `<key>` is the stage's first dispatch key — `visual-verify`, or the one
the fix procedure (`skills/flow/fix.md`) gives a round's re-run — and `prepare` suffixes it per
**The verifier dispatch** (`skills/flow/visual-verify.md`). `--worktree` names every other worktree of
this run's resolved set, the same set **Verify** above resolved. `--merge-base` passes each worktree's
own merge base from this run's working notes. `--app` passes each app the stage may probe, the
`<app-root>=<resolved-url>` argument the pre-flight check takes, prefixed with its worktree.
A worktree held in a shell variable is written braced, `--app "${W}:<app-root>=<resolved-url>"`
— zsh reads `$W:s…` as a history modifier.

1. **Resolve the section — with the guard, not by eye.** `prepare` pipes each worktree's diff into
   `flow-guard check-visual-trigger <worktree>`; exit 2 with a `VISUAL-TRIGGER-NOT-CONFIGURED:` line on stderr
   skips that worktree for the rest of this stage.
2. **Match the diff** — the same call's verdict. Exit 0 → at least one changed path matched a declared
   `ui paths` glob; continue. Exit 1, or exit 2 with a `VISUAL-TRIGGER-CANNOT-ANSWER:` line, skips the
   worktree.
   `flow-guard check-visual-trigger` owns the glob semantics (`**` spanning directories, a leading
   dot-slash prefix, an absolute glob, a glob with a space); nothing here restates them.

   In a run whose resolved set holds more than one worktree, every other worktree's diff is also
   matched against the canonical worktree's `ui paths` through `flow-guard check-visual-trigger --repo
   <repo-name>`, each path prefixed with that worktree's repository name — its main checkout's
   basename, the name a glob such as `gymie-frontend/**` spells another repository's files with — and
   its exit 0 makes the canonical worktree survive this step as the canonical worktree's own exit 0
   does.

   Each worktree's diff runs from its `--merge-base` — never `origin/<base>` or another ref picked by
   hand, which matches every path the base gained since the worktree was cut as this change's own.

   **On a fix run, both calls read the fix run's own diff.** Each worktree is passed
   `--since <worktree>=<its fix-start sha>` (`skills/flow/fix.md`), so a fix run whose delta touches no `ui paths` skips this
   stage at step 2 as exit 1 does.
   No worktree is passed `--since` when the change's newest verifier dispatch row (the newest row `flow record dispatches -change <name>` returns with `role` `verifier` and a `key` starting with `visual-verify`) does not have `outcome` `completed` — an earlier run's UI change that was never verified may lie before the fix-start sha.

**A `skipped` decision ends the stage at step 2.** When this run's decision
(`<abs-worktree>/.superpowers/sdd/decision.json`) records `visual.verify` `skipped` (step 6 of
**Plan check**, `skills/flow/brainstorm.md`), `prepare` closes the stage `skipped` and prints
`NEXT: skip — skipped at plan check — <reason>`, and every worktree surviving step 2 prints
`Visual: skipped at plan check — <reason>`.
`required`, `not configured` or no `visual` at all takes the load below, unchanged.

**`NEXT: skip — <why>` closes the stage and dispatches nothing**, and loads nothing.
**Load `skills/flow/visual-verify.md`** on any other closing line, and act on that line
there — it carries **The verifier dispatch**, steps 3–13, this stage's `## Report`, **Blocking**, and
this stage's `end` mark.

## Stage, excluding the planning paths

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

> **`<project>/spectre/changes/` is never part of a task commit.** This step only confirms nothing
> slipped in.

**Load `skills/flow-contracts/git-boundaries.md`** before committing below.

**The PR-branch split.** Every task and fixup commit already sits on the branch,
pushed as it landed (**Branch backup**, `skills/flow-contracts/git-boundaries.md`). If the state
file records a `prUrl`, a PR is already open, so this run also commits whatever the operator
edited at the human gate, and pushes
everything to the PR branch; otherwise this step commits nothing more.

**Load `skills/flow-contracts/git-boundaries-commit-chain.md`** only when the state file records a
`prUrl`.

On that path only — and in this order — run
`flow record render -change <name> -kind all -repo <canonical-worktree>` (the same member the
ledger render above targets); then that chain in `<worktree>`, `<impl-msg>` its
message; then push the branch
`--force-with-lease`, since the split reshaped it. `<impl-msg>`
covers working-tree edits the operator made at the human gate without staging them — derive it the
same way a fixup commit's subject is derived — `fix(<module>): <what changed since the last task
commit>`.

The render overwrites in place. **A non-zero exit means a destination was refused or could not be
written** — report it, and continue committing the fix.

## Write `IN_PROGRESS`

**Append this run's own narrative first.** Append to `narrative.md` in the change's directory
under `<abs-worktree>` (**A change's directory**, `skills/flow-contracts/pipeline.md`) (create it
with the title `# <name> —
session narrative` when absent) one section `## <YYYY-MM-DD> — <creating run | fix run>` holding
this session's own prose account of the run — problems hit, workarounds, time sinks, environment
gaps, operator decisions taken mid-run, every approach tried and abandoned with why it failed —
and nothing the ledger or panel record already holds. The abandoned approaches and their failure
reasons are recorded because the diff shows only the winning shape — a deferred self-review pass
reads the narrative, and the negative results live nowhere else in the change's record. The narrative itself is gitignored, written on disk and nothing else; nothing stages it.

**The hooks have ended every row whose agent stopped, bar one whose final message they could not read yet**; before the state is
written, list the rows still open:

```bash
flow record dispatches -change <name> -C <canonical-worktree> \
  | jq -r '.[] | select((.outcome // "") == "") | "\(.key) \(.sessionToken) \(.agentId // "inline")"'
```

**A row whose agent returned is never aborted**. A row with an agent id whose transcript (`<session folder>/subagents/agent-<agent id>.jsonl`) exists is closed from that transcript, which is complete by now: `flow-guard end-dispatch-row <that transcript> <session id> <canonical-worktree>`, which prints `END-DISPATCH-ROW-OK` and ends the row on its real final message. Only a row that guard prints `END-DISPATCH-ROW-NA` for, or that has no transcript, is ended `-outcome aborted` — its agent is gone:

```bash
flow record dispatch end -change <name> -C <canonical-worktree> -key <key> -session-token <its token> -outcome aborted
```

**Record the `**Summary:**` bullets next**, per **Summary and live-stack line, before every handoff** (`skills/flow-contracts/pipeline.md`).

**Every run writing `IN_PROGRESS` then runs the unfinished-work guard on itself**, a creating run
as well as a fix run, once per worktree of the resolved
set, with the arguments **Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`) passes it —
`flow-guard check-unfinished-work <worktree> <name> <canonical-worktree> <worktree>=<merge-base>...`. An `OUTSTANDING:` line is
this run's own unfinished work: finish what its breakdown names — a missing inline verify is
run and recorded, never recorded after the fact for work that did not run — and run the guard
again, per **Finish it — never a prompt** (`skills/flow/unfinished-work-gate.md`), whose false-positive record covers an item no change can finish; the state is written once every worktree prints `CLEAR:` or every remaining item is so recorded — never a stop or a question. A missing verdict line or an
exit 2 is reported and stops the run at the state it read.
The guard finds a
code commit no later completed `verify` row covers unfinished, so an unrecorded inline verify stops this run too.

Write the state file: `IN_PROGRESS` from `STARTED`, otherwise **the state exactly as read**.
`worktrees` should already carry one absolute-path key per affected worktree and its merge base,
so this step re-reads the current record and
confirms every resolved worktree is present rather than reconstructing the map from scratch;
add any entry still missing (a worktree added after the last incremental write) before
proceeding. Carry `artifactUrl` (always `null` under `/flow`), `jiraIssue` and `prUrl`
forward verbatim, and any legacy `planningEffort` or `models` an older record holds.
The state file lives outside the repo — never `git add` it.

```bash
flow stage end -command '/flow' -stage flow.verify -outcome completed <name>
```

**Then land straight away unless a question is open** — a run that lands re-records first any
`**Summary:**` bullet changed after the guard, since no printed block records it again. A run that
leaves a `?` line for its `Decisions:` prints the block below and ends the run: an open question
the operator must answer is never landed past. Any other run prints no block but its 🤖
`Decisions:` lines, when it took any, under a `**Decisions:**` heading — integrate's handoffs carry
no such field — and continues at **Deciding which run this is** (`skills/flow/integrate.md`) in the
same turn, as if a bare `/flow <name>` had started it; that run prints its own handoff or stop.
This is the one chain into integrate **Every invocation is re-entrant**
(`skills/flow-contracts/pipeline.md`) sanctions.

````
**Summary:**
- <what this round changed for a user of the app, in product terms — a few bullets at most>
- <a line another contract requires in the handoff — a caveat, a manual check, a push command, an open defect — one bullet each>

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

**The `Decisions:` lines name every prompt this run answered itself** on its recommended
option (**Auto-resolution**, `skills/flow-contracts/operator-prompts.md`), so the operator can
overrule any of them with a fix run; it reads `none` when the run took none and left no open
question. Planning answers the `## decisions: recommended` mode took are named here the same way.

**The parent prints this block directly, as this stage's own output** — no return, no relay: the
running session assembles it here and shows it to the operator in the same turn (**The parent
orchestrates directly**, `skills/flow/implement.md`).

## Guardrails

- **Never** create a second worktree for the same change.
