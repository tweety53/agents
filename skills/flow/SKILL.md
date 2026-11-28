---
name: flow
description: Single-command pipeline — brainstorm, implement behind the decided review panel, then integrate and clean up across the same three-state pipeline, pausing only at the human gates. Re-run to resume, fix, or integrate. Use for /flow.
allowed-tools: Bash(spectre:*), Bash(flow:*)
license: MIT
---

Drive the three-state pipeline (`STARTED` → `IN_PROGRESS` → `FINISHED`) end to end in one command.
`/flow` is one command, one state file, three states, with its content organized by topic. This
file is the router: it resolves state and dispatches into the topic file that owns the phase in
force. Nothing here duplicates that file's own content.

**Announce at start:** "Using flow for change `<name>`."

**Load `skills/flow-contracts/pipeline.md` first** — canonical for the three states, the
transition table's shape, stage-mark mechanics, the guard-presence check, guard resolution, the
handoff shape and change-name resolution. Its **State transitions** table is `/flow`'s contract;
**Stage keys** (`skills/flow-contracts/pipeline.md`) names which phase file marks each key.
Load every contract file with the harness's file-read tool — Read on Claude Code — in offset
chunks wherever one read cannot return the whole file, never with `cat`: `pipeline.md` and
`finish-contract-run1.md` exceed the shell tool's output cap, and a `cat` load arrives truncated.

**No flags.** The only argument is the optional change name/description on a creating or resuming
run, or fix instructions at `IN_PROGRESS`; report anything else rather than ignoring it.

**One exception: `--base <branch>` on a creating run** names the branch the change is cut from and
lands on, in place of the one `origin/HEAD` points at. Kickoff records it on the change's branch in
every repository (`flow-guard kickoff-worktree <project> <name> <base>`, and
`flow-guard kickoff-worktree --from <this change's worktree>` for every worktree added after kickoff), first creating it from that repository's default branch where its origin lacks it, and `flow-guard resolve-base-branch` reads
it back for every later phase, so it is never passed again.
A recorded base gone from origin — merged and deleted — is refused by name by `flow-guard resolve-base-branch` (`<agents repo>/scripts/flow-guard resolve-base-branch`): re-point every worktree of the change at the branch that absorbed it with the three commands integrate's `STOP: base-mismatch` answer names (`skills/flow/integrate.md`), then re-run.

## Model resolution

**The implementer groups, the review panel and visual verification are decided per change, never
configured** — by the headless plan check at **Plan check** (`skills/flow/brainstorm.md`) — and the
run reads them from the recorded `decision.json`.

**Every dispatch runs on its role's pair, the decision's for a pass-1 bundle, else
picked by `flow-guard pair-judge` at dispatch time** — **Model and effort**
(`skills/flow-contracts/model-policy.md`) is canonical.

**A plain-language session instruction overrides the pair(s) it names for this run only** —
"use opus for the panel", "implement on sonnet" — per **Model policy**
(`skills/flow-contracts/model-policy.md`).
The override never rewrites the settings store; the dispatch hook records the pair each dispatch ran on.

## Reading the state

```bash
flow state get <name-or-best-guess> -C <repo-root>
```

Run it as its own Bash call, never chained to another command. **No state** is exit 1 with
`flow: no state recorded for <project>/<name>` on stderr; an exit 1 without that line is a failure
to report, never a creating run.

- **No state** as above, or exit 0 with `"synthetic": true`: a creating run. See
  **A. Resolve the change and write `STARTED`** (`skills/flow/brainstorm.md`); the run ends at
  the plan gate with a `/clear` handoff, and the next `/flow <name>` enters **The parent
  orchestrates directly** (`skills/flow/implement.md`).
- **Exit 0, `"state": "STARTED"`** — a creating run interrupted before it reached `IN_PROGRESS`. See
  **Resuming at `STARTED`** (`skills/flow/resume.md`).
- **Exit 0, `"state": "IN_PROGRESS"`, an argument other than the change's name** — a fix run — or a plain message, per
  **A plain message at IN_PROGRESS** below. See
  `skills/flow/fix.md`; unless the run leaves an open question, the run continues into
  integrate once it writes the state (**Write `IN_PROGRESS`**, `skills/flow/verify-and-handoff.md`).
- **Exit 0, `"state": "IN_PROGRESS"`, no argument or the change's name alone** (**Change name
  resolution (all `/flow*` commands)**, `skills/flow-contracts/pipeline.md`) — an integrate run. See
  **Deciding which run this is** (`skills/flow/integrate.md`).
- **Exit 0, `"state": "FINISHED"`** — emit the wrong-state handoff from **Wrong state for this
  command** (`skills/flow-contracts/pipeline.md`): the change has landed. Proceed only on an
  explicit override.

### A plain message at IN_PROGRESS

**Trigger.** A message with no `/flow` invocation that reports a problem or asks for a change to
the change's code or artifacts, in a session whose most recent `/flow` run marked `<name>`. `<name>`
comes from the `flow: this session's last /flow run marked change <name>` context line
`flow-guard hook flow-active-change` injects, or from this session's own context when no hook is
registered.

**Gate.** `flow state get <name>` must answer `IN_PROGRESS`. Any other answer, or an unreachable
store, makes the message an ordinary turn, never a wrong-state handoff, since nothing was invoked.

**Nothing is asked.** The session decides: a message that reads as a change request is run as a
fix, and one that reads as a question is answered as an ordinary turn.

**The run.** Announce `Using flow for change <name> — fix run from a plain message`, and continue at `skills/flow/fix.md` with the verbatim message
as the fix instructions — nothing else about the fix run changes.

**Not a trigger.** A question, a discussion, an unrelated task, a session that never ran `/flow`,
and a Jira key named in prose without the slash.

**Check guard presence.** Per **Guard presence check** (`skills/flow-contracts/pipeline.md`),
confirm `flow-guard` resolves on `PATH`.

**The `<change>` argument to every mark below is always a resolved change name.** On a creating run
the name does not exist until **A. Resolve the change and write `STARTED`**
(`skills/flow/brainstorm.md`) produces it — defer every mark until that
section has produced it, per **Change name resolution (all `/flow*` commands)**
(`skills/flow-contracts/pipeline.md`). This router reads state above using
a guess or the best available name, which is legal for a read; it is never legal for a mark.

**No mark or record carries a session token**: the CLI infers the calling agent's, per **Stage
marks** (`skills/flow-contracts/pipeline.md`).

## Full-suite instruction

**A plain-language instruction to run the full suite, given at any point of any `/flow` run of
the change, covers one run and is stored nowhere.** Given before the run reaches a `test` list,
the next `test` list it reaches runs as `flow-guard run-verify-list -full <worktree> test`, with the
recording arguments that list carries; given when no `test` list is ahead, the session runs that
command at once in each of the change's worktrees. The command runs only under the operator's own
`FLOW_OPERATOR_FULL_SUITE=1`, which no session sets. It is one line in the run's narrative and one
`Decisions:` bullet. Every later `test` list runs the default one (**Inline verify**,
`skills/flow/verify-and-handoff.md`).

## Guardrails

- Never ask a planning-effort, model, or review-panel question on a creating run. The
  panel is the recorded decision's, never asked; see **Model resolution** above and
  **Review panel** (`skills/flow/review-panel.md`).
- Never skip brainstorming's design gate, or leave `tasks.md` a thin scaffold.
- Never advance the state past what the phase in force is entitled to write — a fix never moves
  the state; an implementation run only ever writes `IN_PROGRESS`; only run 2 (`skills/flow/cleanup.md`)
  writes `FINISHED`.
- Never end a turn mid-run over a phase file's length; read it in offset chunks and continue
  (**A run's size is never a stop.**, `skills/flow-contracts/pipeline.md`).
