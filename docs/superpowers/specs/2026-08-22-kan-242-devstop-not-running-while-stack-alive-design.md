# KAN-242 — `devStop` false negative, and run 2's missing live-process gate

**Jira:** KAN-242 · **Change:** `kan-242-devstop-not-running-while-stack-alive`

Second occurrence of the KAN-129 incident. `./gradlew devStop` reported `not running` for all four
services while four processes were live and holding 9931, 8080, 3000 and 4851. `/myflow-finish`
run 2 then removed both worktrees, taking the `.dev-stack` records that were the only thing able to
name those processes. `devWorkspaceInventory`, run afterwards by chance, was the only control that
caught it.

Two independent defects, in two repositories.

## Scope

One change, change directory in `agents`, two worktrees:

| Repository | Carries | Delta spec |
|---|---|---|
| `agents` | run 2's live-process gate, its guard and test harness | `myflow-finish-cleanup` |
| `gymie` | `devStop`'s not-running inference | **none — see the decision below** |

## Decisions

### The change directory lives in `agents`, and the gymie half ships without a delta spec

**ID:** change-directory-in-agents
**Status:** active
**Chosen:** one change, directory in `agents` — the operator's explicit choice after the cost was
stated. The gymie worktree carries implementation commits and tests only.
**Considered:**
- *Two sibling changes, one per repository, both linked to KAN-242* — each half gets a correct delta
  spec against its own repository's capability set, its own review panel and its own PR. Ruled out
  by the operator in favour of a single change.
- *Change directory in `gymie`* — inverts which half loses its spec. Worse, because the `agents`
  half is the one altering a contract that `/myflow-finish` and `/myflow-status` both consume, so
  it is the half whose spec matters more.

**The cost, stated rather than discovered later:** `agents/openspec/specs/` and
`gymie/openspec/specs/` hold disjoint capabilities, and one change has one change directory. The
`devStop` fix therefore lands in gymie with no delta spec recording the behaviour it establishes.
Nothing in the pipeline will report this as missing.

### A live process blocks worktree removal, rather than being reported after it

**ID:** live-process-is-a-pre-removal-gate
**Status:** active
**Chosen:** a new worktree-cleanup check, run after check 5's `## stop` command and before removal.
While the worktree still exists, the project's own `devStop` can still read its pid files, so the
operator has a remedy. The gate is what verifies the stop took effect — which is precisely what
nothing did.
**Considered:**
- *A seventh registry row in `check-cleanup-complete.sh`, as the ticket suggests* — reports the leak
  only after the worktree and its pid files are gone, leaving killing pids by hand as the sole
  remedy. That is the KAN-242 end state, detected rather than prevented.
- *Both* — strictly more coverage at two implementation sites, where the second fires only when the
  first was skipped. Rejected as complexity nobody asked for.

### Detection is a generic cwd scan, not a project-declared command

**ID:** generic-cwd-scan
**Status:** active
**Chosen:** one `lsof -d cwd -Fpn` pass, filtered to processes whose working directory is at or
under the worktree path. Project-agnostic, requires no project to declare anything, and works for
every project myflow is installed into on the day it lands.
**Measured:** `lsof -d cwd -Fpn` returns 290 cwd rows in a single pass on this machine, no `sudo`,
emitting `p<pid>` / `f cwd` / `n<path>` triples.
<!-- measured: lsof -d cwd -Fpn 2>/dev/null | grep -c '^n' @ /Users/tweety53/Projects/agents, 2026-08-22 -->
**Considered:**
- *A fourth `## workspace isolation` command verb* — precise, but the command table's vocabulary is
  closed at `create`, `remove`, `survivors`, a guard enforces that, and every project would have to
  opt in before the gate protected it.
- *Extending `survivors`* — no new verb, but it stretches that command from "resources inside
  services" to "processes on this machine", and it runs after `remove`, too late for a gate.

**Why cwd and not ports.** A port proves nothing: a live workspace legitimately holds one. What a
removed worktree's process cannot have is a working directory that still exists. `gymie`'s own
`OrphanScan.kt` reaches the same conclusion independently, which is corroboration rather than
coincidence.

### A blocked removal is a hard failure, not a confirmable disclosure

**ID:** hard-block
**Status:** active
**Chosen:** a live process fails the check; every worktree is left alone and run 2 stops at
`IN_PROGRESS`, reporting the pids and the command that reaches them. Matches the existing rule that
any failed check leaves every worktree alone.
**Considered:**
- *Disclosure plus explicit confirm, like check 4* — a prompt that can be approved reflexively, which
  is how KAN-129 and KAN-242 both happened.
- *Hard block with a `/myflow-fast` override* — `/myflow-fast` overrides check 4's ask because the
  records worth keeping are already out of the worktree by then. That argument does not transfer: a
  live process is not a preserved record, and an unattended run is exactly the case that orphans it.

### `not running` must be observed, never inferred from absent records

**ID:** not-running-requires-a-definite-negative
**Status:** active
**Chosen:** `stopService` may print `not running` only on a definite negative from both halves —
`records.pid == Recorded.None` **and** a port that resolved and has no listener. Anything else
becomes a third outcome naming which half was unknown.
**Considered:**
- *Investigate the `lsof`-showed-free contradiction first* — deferred; recorded as an open question
  below. The gate above catches that failure mode whatever its cause.
- *Reproduce before deciding* — the defect is legible in the source without a reproduction, and the
  fix reuses the file's own existing vocabulary.

**The defect, precisely.** `stopService` reaches `not running` when
`trackedPid == null && orphans.isEmpty()`. Both conjuncts are satisfied by absent information:
`trackedPid` is `(records.pid as? Recorded.Value)?.value`, which collapses `Recorded.None` with
`Recorded.Unknown` — the exact collapse `StateFile.kt`'s own documentation warns against — and
`orphans` is empty whenever `sweepPortFor` returns null, which in an isolated workspace it
deliberately does for an unusable port record. So `not running` is a statement about this run's
records, printed as a statement about the process.

## Open questions

### Why did `lsof` on all four ports report them free immediately after the stop?

**ID:** lsof-reported-free
**Status:** open
**Why it is open:** deferred by the operator in favour of fixing the inference first. The ticket
records that `lsof` on 3000/8080/9931/4851 showed all four free right after `devStop`, yet
`devWorkspaceInventory` later found live processes holding those same four ports. Nothing in the
source read so far explains that, and it may be a second, independent defect.
**What it affects:** whether the `devStop` fix is sufficient on its own. If the ports were genuinely
free at that moment, something re-bound them afterwards, and neither this change's gate nor its
inference fix addresses that.

## What is not in scope

- The archive-branch leftovers named in the temporary artifacts registry.
- `pipeline.md`'s size. Raised in session and worth its own ticket; folding it in here would put an
  unrelated contract refactor behind this change's review panel.
