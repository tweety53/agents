---
name: flow
description: Single-command pipeline — brainstorm, implement behind the review panel resolved from the settings store, then integrate and archive across the same three-state pipeline, pausing only at the human gates. Re-run to resume, fix, or integrate. Use for /flow.
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
**Stage keys** (`skills/flow/stage-keys.md`) names which phase file marks each key.

**Then register this run's steps** with the harness's task-list mechanism, before any work begins,
and keep each entry's status current as the run proceeds, per **Progress visibility**
(`skills/flow-contracts/pipeline.md`).

**No flags.** The only argument is the optional change name/description on a creating or resuming
run, or fix instructions at `IN_PROGRESS`; report anything else rather than ignoring it.

## Model resolution

**Resolve this once, near the top of every run, before any dispatch below reads it:**

```bash
SETTINGS_JSON="$(flow settings get 2>/dev/null || true)"
REVIEWERS="$(printf '%s' "$SETTINGS_JSON" | jq -r '.reviewers[]')"
VERIFY_MODEL=opus
```

A non-zero exit from `flow settings get` means the settings store could not be reached — there is
no per-change fallback file for this record. Settings unreachable is never a reason to block
implementation.

**Execution mode, every dispatch's model and effort, and the review panel are decided per change,
never configured.** The plan's class and rolls decide them — **Decide**
(`skills/flow/brainstorm-planner.md`) — and the run reads them from the recorded
`decision.json`.

**`REVIEWERS` resolves from the same call, into the roster a `micro` decision's panel
dispatches** — the one class whose `panel` is the string `default` (`skills/flow/review-panel.md`
is canonical for what dispatching it means):

| Store state | Resolved roster |
|-------------|-----------------|
| Reachable, list non-empty | exactly the list |
| Reachable, list empty | `primary` alone |
| Unreachable | `primary`, `principles` (`DefaultReviewers` in `<agents repo>/stats/internal/store/settings.go`), naming this a fallback rather than a resolved value |

**Every other dispatch runs on its decision pair** — the implementer, every panel slot and the
panel-fix subagent — or, with no pair recorded, on the literal `opus`. **Model and effort**
(`skills/flow/brainstorm-planner.md`) is canonical for that no-pair `opus`, the tooling analyst's
included, and for how each pair's model is chosen.

**A plain-language session instruction overrides the pair(s) it names for this run only** —
"use opus for the panel", "implement on sonnet" — per **Model policy**
(`skills/flow-contracts/model-policy.md`).
Record the instruction with the dispatch it changes; an override nobody wrote down is
indistinguishable from a mistake. The override never rewrites the settings store or the decision's
pairs; it lands in the decision's `overrides`.

## Reading the state

```bash
flow state get <name-or-best-guess> -C <repo-root>
```

- **Exit 1**, or exit 0 with `"synthetic": true` — **no state**: a creating run. See
  **A. Resolve the change and write `STARTED`** (`skills/flow/brainstorm.md`); the run ends at
  the plan gate's **Yes** with a `/clear` handoff, and the next `/flow <name>` enters **The parent
  orchestrates directly** (`skills/flow/implement.md`).
- **Exit 0, `"state": "STARTED"`** — a creating run interrupted before it reached `IN_PROGRESS`. See
  **Resuming at `STARTED`** (`skills/flow/brainstorm.md`).
- **Exit 0, `"state": "IN_PROGRESS"`, an argument present** — a fix run — or a plain message, per
  **A plain message at IN_PROGRESS** below. See
  **3. Documenting a fix, before implementing it** and then **The parent orchestrates directly**
  (`skills/flow/implement.md`).
- **Exit 0, `"state": "IN_PROGRESS"`, no argument** — an integrate run. See
  **Deciding which run this is** (`skills/flow/integrate.md`).
- **Exit 0, `"state": "FINISHED"`** — emit the wrong-state handoff from **Wrong state for this
  command** (`skills/flow-contracts/pipeline.md`): the change is archived — or withdrawn, when
  the record's `withdrawn` field is true (`skills/flow-contracts/state-file.md`); the handoff
  names which. Proceed only on an
  explicit override.

### A plain message at IN_PROGRESS

**Trigger.** A message with no `/flow` invocation that reports a problem or asks for a change to
the change's code or artifacts, in a session whose most recent `/flow` run marked `<name>`. `<name>`
comes from the `flow: this session's last /flow run marked change <name>` context line
`hooks/flow-active-change.py` injects, or from this session's own context when no hook is
registered.

**Gate.** `flow state get <name>` must answer `IN_PROGRESS`. Any other answer, or an unreachable
store, makes the message an ordinary turn, never a wrong-state handoff, since nothing was invoked.

**Ambiguity prompt.** A message that reads as either a question or a change request asks once,
shape per **The shape** (`skills/flow-contracts/operator-prompts.md`): **Run this as a fix of
`<name>`?** — **Yes** *(recommended)* / **No — ordinary turn**. Silence takes Yes and the handoff
carries the ⚠ line.

**The run.** Announce `Using flow for change <name> — fix run from a plain message`, generate this
run's own session token per **Generate this run's session token once** below, and continue at **3.
Documenting a fix, before implementing it** (`skills/flow/implement.md`) with the verbatim message
as the fix instructions — nothing else about the fix run changes.

**Not a trigger.** A question, a discussion, an unrelated task, a session that never ran `/flow`,
and a Jira key named in prose without the slash.

**Check guard presence.** Per **Guard presence check** (`skills/flow-contracts/pipeline.md`),
confirm every guard `/flow` can invoke — every `<name>.sh` a fenced command line or the prose of
`skills/flow/*.md` and the contract files it loads names, the set
`<agents repo>/scripts/check-guard-symlinks.sh`'s rule 2 derives — is present in `<skill-dir>/scripts/`.

Every `flow-guard` shim — every guard whose source loads `<agents repo>/scripts/lib/flow-guard.sh`,
by `$SCRIPT_DIR/lib/…` or by `$(dirname -- "${BASH_SOURCE[0]}")/lib/…` — also requires that file as
a `<agents repo>/scripts/lib/` sibling — the same
sibling-dependency rule `<agents repo>/scripts/check-guard-symlinks.sh`'s rule 2 already applies to
every other guard above.

**The `<change>` argument to every mark below is always a resolved change name.** On a creating run
the name does not exist until **A. Resolve the change and write `STARTED`**
(`skills/flow/brainstorm.md`) produces it — defer every mark until that
section has produced it, per **Change name resolution (all `/flow*` commands)**
(`skills/flow-contracts/pipeline.md`). This router reads state above using
a guess or the best available name, which is legal for a read; it is never legal for a mark.

**Generate this run's session token once, right here, before the first mark any phase file below
makes — a short, unique literal string — and reuse that exact same value at every `stage begin` this
run makes, including inside every phase file it dispatches into.** One run, one token, never a fresh
one per mark or per phase file. `<literal-token>` in each mark is that string written in place of
the placeholder, never the placeholder itself, and a later `/flow` invocation — a resume or a fix
run — generates its own rather than reusing an earlier run's.

## Guardrails

- Never ask a planning-effort, model, or review-panel-roster question on a creating run. The
  roster is the recorded decision's, never asked — the settings store's on a `micro` decision; see **Model resolution** above and
  **Review panel** (`skills/flow/review-panel.md`).
- Never skip brainstorming's design gate, or leave `tasks.md` a thin scaffold.
- Never advance the state past what the phase in force is entitled to write — a fix never moves
  the state; an implementation run only ever writes `IN_PROGRESS`; only run 2 of the archive branch
  and **The withdrawal route** (`skills/flow/brainstorm.md`) write `FINISHED`.
