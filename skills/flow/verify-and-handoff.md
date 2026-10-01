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

**This step does not call the project's `create` command.** `create` is called by whatever starts
the project's applications, per **Project configuration**
(`skills/flow-contracts/project-configuration.md`), and this step starts none of them — it
exports, lints, tests, and hands off.

**After the panel closes, the parent edits no source outside the in-run fix loop.** Source
changes only through **The loop** (`skills/flow/verify-fix-loop.md`) or a fix run the operator
starts; any other source change from here on makes every slot's result stale (**Panel re-runs**,
`skills/flow/review-panel.md`).

**Load `skills/flow/verify-fix-loop.md`** only when this stage's or `flow.visual-verify`'s final
report carries a fixable defect — it carries **The loop** and **The cap**.

### Inline verify

Resolve the commands `project-get.sh <worktree> lint` and `project-get.sh <worktree> test` print
(auto-detect on exit 1), and the known-failures baseline `project-get.sh <worktree> "known failures"`
prints (**Project configuration**, `skills/flow-contracts/project-configuration.md`) — exit 1, the
key absent, is the ordinary case and leaves no baseline for the classification below.
Export the `KEY=value` lines `prepare-workspace.sh` printed for that worktree, then run the lint
commands, then the test commands, in the order printed, then `check-spec-reach.sh <worktree>` —
one more command in the same list, whose exit 0 line `Spec reach: not configured` is the ordinary
case for a project with no `regression checkout` (its header is canonical for its exit codes). Run
every command in order and do not stop at the first failure. **Nothing runs them later** —
`/flow`'s integrate phase has no verification gate — so a non-zero exit blocks this handoff, the
known-failures case of **Inline verify — a failing command** below being the one exception.

```text verified:design.md section 2 of this change
## Report
- `<command>` — exit <n>[ — known failures only]  # the bracketed marker only when the command's failing tests are all known
  <the command's output, verbatim, or its last 20 lines when longer, stated as truncated>
  known failures: <identifier> — <reason>[; …]  # only under an exit line carrying the marker
```

The parent writes this `## Report` itself and shows it as this stage's output.
`prepare-workspace.sh`, both `project-get.sh` calls and this stage's `begin` mark are one Bash
call; the lint and test run, this run's `flow record dispatch begin`, and the ledger render below
are one more.

**Inline verify — a failing command.** When a command exits non-zero, its output is classified
against the known-failures baseline before any attempt is spent on it: a failing test the output
names that a baseline entry matches under the key's own matching rule (**Project configuration**,
`skills/flow-contracts/project-configuration.md`) is a **known failure**. A lint failure is never a
known failure — the baseline matches tests, so only failing-test output is classified. A command
whose output names at least one failing test, and whose every failing test is known, carries the
`known failures only` marker beside its exit line in the `## Report`, names each known failure
under it with its entry's reason, earns **no** inline re-run, and does not block this handoff —
the baseline is the project's own statement that the failure exists on an unmodified tree. Output that names no failing test at all — a compile error, a harness crash — and a
failing test with no matching entry behave exactly as the rules that follow. A failing test with
no matching entry that **the sweep** (`skills/flow-contracts/known-bugs.md`) classifies
pre-existing takes the known-failure course above instead of the re-run below.

A non-zero exit from any command in the list earns **one**
inline re-run of that command — the environmental-flake case. A second non-zero exit from the same
command is the branch's own defect and takes **The loop** (`skills/flow/verify-fix-loop.md`) once
the whole list has run; where the environment itself failed — a missing build prerequisite — it
ends the turn with `## Question` naming the command and its output, verbatim. Never treat a
passing re-run as license to skip the rest of the list — every remaining command in the order
above still runs.

**Recording.** One `dispatches` row per worktree, `-role verifier -key verify -model <parent
model> -effort <parent effort> -agent-id inline`, suffixed `-<worktree basename>` when this
run's resolved set holds more than one worktree — the same convention
**The verifier dispatch** (`skills/flow/visual-verify.md`) uses for the verifier's rows. `begin` is
recorded before the first command in the list; `end` after the `## Report` is written, carrying
`-outcome completed`, or `-outcome blocked -cause <cause>` when a command failed twice —
`test-failure` for a command of the branch that failed twice, `environment` where the
environment itself failed twice (a missing build prerequisite).

**Load `skills/flow-contracts/session-records.md`** before reading the render outcome below.

**Confirm this run recorded a ledger** — rendering into the canonical worktree, the member of
this run's resolved set whose own `<project>/<spec-root>/changes/<name>/tasks.md` exists (the
same member **1. Check for unfinished work**, `skills/flow/integrate.md`, passes
check-unfinished-work), never into whichever worktree this pass runs in, so a multi-worktree run
writes one copy, not one per worktree:

```bash
flow record render -change <name> -kind ledger -repo <canonical-worktree>
```

Read the outcome word, not the exit code. `rendered: <dest>` is ordinary. **`MISSING: ledger — no
rows for <name>` means this run recorded no dispatch at all**, reported plainly here rather than
discovered later. The outcome words are the table under
**Rendering the session records** (`skills/flow-contracts/session-records.md`).

```bash
flow stage end -command '/flow' -stage flow.verify -outcome completed <name>
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

   Exit 2 with a `VISUAL-TRIGGER-NOT-CONFIGURED:` line on stderr → this worktree prints
   `Visual: not configured` and is skipped for the rest of this stage.
2. **Match the diff** — the same call's verdict.
   Exit 0 → at least one changed path matched a declared `ui paths` glob; continue. Exit 1 → this
   worktree prints `Visual: no UI paths touched` and is skipped for the rest of this stage. Exit 2
   with a `VISUAL-TRIGGER-CANNOT-ANSWER:` line → report its stderr and skip this worktree the same
   way exit 1 does.
   `check-visual-trigger.sh` owns the glob semantics (`**` spanning directories, a leading
   dot-slash prefix, an absolute glob, a glob with a space); nothing here restates them.

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
<name> "<impl-msg>" "chore(spectre): plan"`; then push the branch
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

Resolve the run instructions for the handoff's `Running:` section. It writes no file.

- **Every start command comes from `<project>/.flow/project.md`'s `## run`**, with every path
  made absolute.
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
  name — the ones just resolved, and no others — from the project's `## run` commands.

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

- **The stack behind the URLs is checked, not trusted.** When the `Running:` block below would
  carry URL lines, run `check-dev-stack-fresh.sh <worktree>` first: the project's declared
  `fingerprint` row (the `## visual verification` section) proves what the stack serves is the
  worktree's own build. The guard's own header is canonical for what its three exits
  cover. Exit 0 adds nothing. Exit 1 adds one line beside the URLs, naming the application, its
  resolved URL and the guard's own stderr reason, restart-shaped and never a restatement of the
  verdict: `Stale: <app> (<url>) — <the guard's reason>; restart the stack before testing.`
  Exit 2 adds `Freshness: unverified — <the guard's stderr reason>` instead, the same
  visible-gap rule the visual-verification procedure's own step 6 (`skills/flow/visual-verify.md`) runs on a missing row. A refused start
  (below) already ends the run, so this check only ever runs on a start that succeeded or was
  skipped by the protected-service rule.

- **The `Running:` block is the start command's own output.** Once the start above succeeds, its
  `Running:` lines are every line of that command's output containing `http://` or `https://`,
  verbatim, followed by the `## stop` command (or, where `## stop` declares none, the same `## run`
  command that started it). `devStart`-style commands already announce every resolved URL on
  success; that announcement — not the isolation table, which cannot know the pinned-port handoff
  stack's URLs once a port has moved — is the source.

- **A refused start is relayed, never resolved.** When the start command above exits non-zero
  because a port it needs is already held, this stage does not retry, does not pick another port,
  and does not stop the holder. It prints that command's output verbatim, then the holder's own
  `devStop`-style stop command when the output names the holder's worktree (or, when it does not,
  a note to run `check-worktree-processes.sh`-style diagnosis and stop the holder's stack by hand).
  This run then ends at `IN_PROGRESS` with the diff still staged, and the handoff names `/flow
  <name>` as the next command — the same change, re-run once the holder is stopped. It never stops
  a stack this run did not start.

```bash
flow stage end -command '/flow' -stage flow.run-instructions -outcome completed <name>
```

## Write `IN_PROGRESS`

```bash
flow stage begin -command '/flow' -stage flow.write-in-progress -harness <harness> -session-token mf-<literal-token> <name>
```

**Append this run's own narrative first.** Append to
`<abs-worktree>/spectre/changes/<name>/narrative.md` (create it with the title `# <name> —
session narrative` when absent) one section `## <YYYY-MM-DD> — <creating run | fix run>` holding
this session's own prose account of the run — problems hit, workarounds, time sinks, environment
gaps, operator decisions taken mid-run — and nothing the ledger or panel record already holds. The
write-in-progress planning commit carries it (**Planning commits**,
`skills/flow-contracts/git-boundaries.md`), made once the append lands; nothing else stages it.

Write the state file: `IN_PROGRESS` from `STARTED`, otherwise **the state exactly as read**.
`worktrees` should already carry one absolute-path key per affected worktree and its merge base,
so this step re-reads the current record and
confirms every resolved worktree is present rather than reconstructing the map from scratch;
add any entry still missing (a worktree added after the last incremental write) before
proceeding. Carry `artifactUrl` (always `null` under `/flow`), `jiraIssue`, `planningEffort`
(never written — `null` unless a legacy record holds a level), `models.default` (a legacy field,
likewise never written — `/flow` resolves models from the settings
store per run, never records a value into the per-change state) and `prUrl` forward verbatim.
The state file lives outside the repo — never `git add` it.

```bash
flow stage end -command '/flow' -stage flow.write-in-progress -outcome completed <name>
```

**Produce the handoff's `Records:` and `Costs:` lines with one call**, `-C` the canonical worktree the ledger render above targets, one `-worktree` per other
affected worktree:

```bash
flow record handoff-lines -change <name> -C <canonical-worktree> [-worktree <abs-worktree> ...]
```

It exits 0 always; render each line exactly as printed.

```
## Implementation staged — review and test | Implementation committed — review and test

**Change:** <name>
**Panel:** clean — roster: <the slot list this run dispatched>; reduced: <"docs-only — " or "late-fix — " followed by the resolved slot(s) not dispatched, or "no">; <default|decided — class, compact?, rerun policy, dispatches: <group> · <group>, rerun: <model>/low>; added this run: <slot(s) an explicit operator instruction added beyond the resolved list, or "none — resolved list ran alone">
**Visual:** not configured | no UI paths touched | pre-flight failed — <the failing checks and their evidence> | open after 5 in-run fix rounds — <defect>[; <defect> …] | <view>: <absolute screenshot path>[, <view>: <absolute screenshot path> …][ — push with: git -C <regression checkout> push]
**Tooling analysis:** none — no miss | <absolute path of sweeps-<n>.md> — <k> sweeps, fold into step 10 of skills/flow/visual-verify-verifier.md? | aborted — verified without added sweeps
**Staged:** N/N tasks staged and uncommitted | N/N tasks committed on branch | committed, plus one planning-artifacts commit, and pushed to the PR branch
**Records:** all writes reached the store | N write(s) journalled — the store was unreachable | unknown — the journal could not be counted
**Costs:** <the line `flow record handoff-lines` printed>
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)
**Auto-resolved:** none | ⚠ <question> → <the recommended option taken>[; ⚠ <question> → <option> …]
**Jira description (pre-edit):** <the text as it stood before the write, verbatim in a fenced block, inside <details> when long> | omitted — this run wrote no description

Worktree:   <absolute worktree path>

Running:
  <url line>  # <from the start command's output>
  <stop command>

Review the diff, then run it:
  git -C <absolute worktree path> diff <merge base>..HEAD
  open -na "IntelliJ IDEA" --args "<absolute worktree path>"

Re-run this command to fix anything you find, or bare to move on to integrating it.

Next:
/clear
/flow <name>
```

**Heading and `Staged:` line select from whether the worktree carries any commits yet.**
Implementation commits per task, so an ordinary run reads "committed"; the "staged and
uncommitted" spelling covers a run resuming before any task committed. **The `Panel:` line is
`/flow`'s own** — it states what **Review panel** (`skills/flow/review-panel.md`) actually
dispatched this run: the resolved roster, its docs-only reduction to `primary` (**The docs-only
reduction**, `skills/flow/review-panel.md`), or its late-fix reduction (**The late-fix
reduction**, `skills/flow/review-panel.md`); any slot an explicit operator instruction added
beyond the resolved list; and, per **Bundled dispatch** (`skills/flow/review-panel.md`) and the `## Decision` block
of **Decide** (`skills/flow/brainstorm-planner.md`), whether the decision's panel was
`default` (a `micro` class) or decided, the run's class, whether the roster was `compact` or `full`, its rerun
policy (`delta`), the dispatch groups as `+`-joined roles and, on a decided panel, the
rerun pair — the same fields and
shape `skills/flow-contracts/handoff-blocks.md`'s `Panel:` line carries for `/flow-status`'s
regenerated view of the same state.

**The `Auto-resolved:` line names every prompt this run answered itself** on its recommended
option (**Auto-resolution**, `skills/flow-contracts/operator-prompts.md`), so the operator can
overrule any of them with a fix run; it reads `none` when the run took none. Planning answers
the `## decisions: recommended` mode took are named here the same way.

**The `Visual:` line reports `flow.visual-verify`'s own outcome.** **Its push clause appears only when step 12
committed to a `regression checkout`.**

**The `Tooling analysis:` line reports **A missed defect — the tooling analysis**
(`skills/flow/visual-verify.md`)**, one entry per worktree whose fix run dispatched one. A
completed analysis names its sweeps file by absolute path, so the operator can decide whether its
sweeps join step 10 for every later change; the run itself never edits `visual-verify-verifier.md`.

**The parent prints this block directly, as this stage's own output** — no return, no relay: the
running session assembles it here and shows it to the operator in the same turn (**The parent
orchestrates directly**, `skills/flow/implement.md`).

## Guardrails

- **Never** create a second worktree for the same change.
