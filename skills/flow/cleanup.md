# Clean up (run 2)

Loaded either by `skills/flow/integrate.md`'s merge-and-push route, in the same invocation, or by a
fresh bare `/flow <name>` invocation once `check-finish-preflight.sh` returns `RUN2` from every
worktree.

**`skills/flow-contracts/finish-contract-run2.md` is canonical for the full procedure.**
`integrate-change.sh` runs it — chained from `land` on the merge-and-push route, or called once
`prepare`'s `NEXT:` names it — with its five stage marks; the parent handles each stop, makes the
Jira transition, and prints the handoff. **Load `skills/flow-contracts/finish-contract-run2.md`
only when** a stop below cites one of its sections, and **load
`skills/flow-contracts/artifacts-registry.md` only when** a `LEFTOVER:` line names a row of it.

Call it with the parent's own shell outside every worktree in the set, per check 6 of **Worktree
cleanup** (`skills/flow-contracts/finish-contract-run2.md`).

```bash
integrate-change.sh cleanup <main-checkout> <name> --harness <harness> --session-token mf-<literal-token>
```

Run it with a Bash `timeout: 600000`: a refresh that moved `<agents repo>/stats/` restarts flowd inside the call,
and a kill partway leaves it down.

- **`STOP: not-merged`** → this is not run 2; fall back to `skills/flow/integrate.md` and **remove
  nothing**.
- **`STOP: disclose`** → relay its `UNCLASSIFIED:` and `DISCLOSE:` lines per check 4 and the
  wave-group copies of **Worktree cleanup** (`skills/flow-contracts/finish-contract-run2.md`), ask
  only what they ask, then re-run `cleanup` adding `--proceed <main-checkout>` for the repository
  the stop names, every earlier `--proceed` kept.
- **`STOP: worktree-cleanup`** → a `REFUSED:` or `HELD:` line, or the guard's exit 2: stop at
  `IN_PROGRESS`, leave every worktree alone, and report each line per **Worktree cleanup**
  (`skills/flow-contracts/finish-contract-run2.md`).
- **`STOP: leftover`** → the verdict rows of step 4 of **Run 2 — the branch is merged**
  (`skills/flow-contracts/finish-contract-run2.md`): print the incomplete block below.
- **`STOP: state`** → report it; the change stays at `IN_PROGRESS`.

A `WORKSPACE-REMOVE-FAILED:` line is **the one exception to the stop-at-the-first-failure rule**:
the script carries on to the cleanup check, which decides the verdict from the project's survivor
report, never from this command's exit code. Relay every clause a `COMPLETE:` line carries after
` — ` word for word, a `SKIPPED:` one included, and each `REMOTE-*` line as the handoff's
**Remote branch:**.

**Load `skills/flow-contracts/jira-integration.md`.** **Transition the issue to Done** on the
`JIRA: transition <KEY> to Done` line, after the state write, per **Jira integration**
(`skills/flow-contracts/jira-integration.md`). A run that stopped at step 4 transitions nothing.

A `REFRESH-REFUSED` line goes into the handoff verbatim, per step 6 of **Run 2 — the branch is
merged** (`skills/flow-contracts/finish-contract-run2.md`).

```
## Finished

**Change:** <name>
**Archived:** <project>/spectre/changes/archive/<name>/ as an absolute path (landed on <base> with the change)
**Worktrees:** removed | left alone — <reason>
**Remote branch:** deleted | already gone | not deleted — <reason>
**Cleanup:** verified
**Self-review:** <project>/docs/self-review/<name>-self-review.md as an absolute path
**Main checkout:** fast-forwarded | already current | <the REFRESH-REFUSED line>
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)
**Jira:** <KEY> → Done | none linked | ⚠ Jira: skipped — <reason>
**Learning:** <the learning note's absolute path> | NONE — <reason>
**Prompt-tuning:** <each case's `index.html` absolute path in `<agents repo>`'s main checkout> | NONE — <reason>
```

**`Learning:` and `Prompt-tuning:` are never omitted** — this block, and `/flow-fast`'s landing report, which carries the same two lines, are where they are canonical. `Learning:` names the file **Learning**
(`skills/flow-contracts/learning.md`) places at `<project>-learning/<name>.md` when it exists; its <!-- citations-guard:allow -->
`NONE` reason is the setting off or the store unreachable. `Prompt-tuning:` names every case the
self-review's step 2a wrote (`skills/flow-self-review/SKILL.md`); its `NONE` reason is angle 8's
own empty-angle sentence.

A run 2 that **completes step 5** is terminal and names **no** next command.

**The Finished handoff, and the run's final report beside it, carry no open item** — nothing
"still open", unverified, untraced or left as the operator's call. Each one is resolved before the handoff prints: fixed,
disproved with its evidence, or a `big` finding filed or declined at step 4's prompt of
`/flow-self-review` (`skills/flow-self-review/SKILL.md`).

On a leftover — or on no verdict at all — the run stops at step 4 instead:

```
## Cleanup incomplete — not finished

**Change:** <name>
**Remaining:** <what the guard named> | unverified — <what the guard reported on stderr>
**State:** IN_PROGRESS — FINISHED is not written while anything remains

<what the operator must clear>

Next:
/flow <name>
```

## Worktree cleanup

Run `remove-change-worktrees.sh` once per repository per **Worktree cleanup**
(`skills/flow-contracts/finish-contract-run2.md`), canonical for its checks, exit codes and relay:

**Check 4 asks only about an irreplaceable, unpreserved entry**, per check 4 of **Worktree cleanup**
(`skills/flow-contracts/finish-contract-run2.md`), canonical for the buckets and that one ask;
everything else it finds is reported and the removal proceeds. **Checks 1, 2, 3, 5 and 6 remain
gates.** Check 6, the live-process check, is named explicitly because it is the one check 4's
proceed-without-asking could plausibly be read as reaching: a live process is not a preserved
record, so `HELD:` and the guard's exit 2 both stop `/flow`.

## Guardrails

- **Never** `git add` the state file, and never move it into the archive.
