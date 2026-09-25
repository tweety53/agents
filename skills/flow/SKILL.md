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
**Stage keys** below names which phase file marks each key.

**Then register this run's steps** with the harness's task-list mechanism, before any work begins,
and keep each entry's status current as the run proceeds, per **Progress visibility**
(`skills/flow-contracts/pipeline.md`). One entry per brainstorming checklist item and artifact on
the planning branch, one entry per `tasks.md` item on the implementation branch, one entry per step
of whichever finish run is executing on the integrate/archive branch — the same granularity
`/flow-fast` used for the branch it is running, since `/flow` runs the same branches.

**No flags.** The only argument is the optional change name/description on a creating or resuming
run, or fix instructions at `IN_PROGRESS`; report anything else rather than ignoring it.

## Stage keys

Every stage `/flow` marks uses a `flow.*` key.

The full key list, in the order each phase file marks them:

| Phase file | Keys |
|------------|------|
| `skills/flow/brainstorm.md` | `flow.kickoff`, `flow.brainstorm`, `flow.design-approval`, `flow.create-artifacts`, `flow.writing-plans`, `flow.decide` |
| `skills/flow/implement.md` | `flow.load-context`, `flow.isolate-workspace`, `flow.document-fix`, `flow.sdd-tdd` |
| `skills/flow/review-panel.md` | `flow.review-panel` |
| `skills/flow/review-panel-optional-slots.md` | `flow.review-panel` — loaded only for a round whose roster carries `bugbot`, `mutation` or an `exp-` slot |
| `skills/flow/verify-and-handoff.md` | `flow.verify`, `flow.visual-verify` (steps 1–2), `flow.stage-diff`, `flow.run-instructions`, `flow.write-in-progress` |
| `skills/flow/visual-verify.md` | `flow.visual-verify` from step 3 — loaded only when a worktree's diff matched a `ui paths` glob |
| `skills/flow/integrate.md` | `flow.preflight`, `flow.unfinished-work-gate`, `flow.landing-question`, `flow.preserve-sessions`, `flow.commit-two`, `flow.landing-routes` |
| `skills/flow/archive.md` | `flow.verify-merge`, `flow.sync-archive`, `flow.commit-archive`, `flow.cleanup`, `flow.verify-cleanup`, `flow.write-finished`, `flow.self-review`, `flow.push-archive` |

## Model resolution

**Resolve this once, near the top of every run, before any dispatch below reads it:**

```bash
SETTINGS_JSON="$(flow settings get)"
DEFAULT_MODEL="$(printf '%s' "$SETTINGS_JSON" | jq -r '.defaultModel')"
REVIEWERS="$(printf '%s' "$SETTINGS_JSON" | jq -r '.reviewers[]')"
VERIFY_MODEL=opus
```

A non-zero exit from `flow settings get` means the settings store could not be reached — there is
no per-change fallback file for this record. Report the CLI's stderr and fall back to the literal
`opus` (the store's own no-row default, per `<agents repo>/stats/internal/store/settings.go`'s `DefaultModel`),
naming that this is a fallback rather than a resolved value, and continue: settings unreachable is
never a reason to block implementation.

**Execution mode, implementer effort and the review panel are decided per change, never
configured.** The plan's class and rolls decide all three — **Decide**
(`skills/flow/brainstorm-planner.md`) — and the run reads them from the recorded
`decision.json`.

**`REVIEWERS` resolves from the same call, into the roster a `micro` decision's panel
dispatches** — the one class whose `panel` is the string `default` (`skills/flow/review-panel.md`
is canonical for what dispatching it means):

| Store state | Resolved roster |
|-------------|-----------------|
| Reachable, list non-empty | exactly the list |
| Reachable, list empty | `primary` alone |
| Unreachable | `primary`, `principles` (`DefaultReviewers` in `<agents repo>/stats/internal/store/settings.go`), naming this a fallback rather than a resolved value — the same pattern as `DEFAULT_MODEL`'s |

**`SELF_REVIEW_MODEL` is not resolved here.** It governs no dispatch and only the archive-phase
self-review pass reads it, so it resolves there, at its point of consumption —
`skills/flow/archive.md` step 9, canonical for it. No run that stops before archive pays for it.

**`VERIFY_MODEL` governs the one verifier dispatch** — `flow.visual-verify`'s (**Visual
verification**, `skills/flow/verify-and-handoff.md`); `flow.verify` runs inline in the parent
and dispatches no verifier. `VERIFY_MODEL` is the fixed literal `opus`, dispatched at effort
`low` through `subagent_type: flow-low`, read from neither the settings store nor
`<project>/.flow/project.md`; a plain-language session instruction does not override it; and it
never falls back, because it is never resolved — the point is a predictable model for mechanical
verification runs regardless of what `DEFAULT_MODEL` resolved to.

**`DEFAULT_MODEL` is the model for all three roles this run dispatches on** — the implementer
(`skills/flow/implement.md`), every panel slot, Bugbot and Security included (every one a
prompt-driven role, per **The roster**, `skills/flow/review-panel.md`),
and the panel-fix subagent (`skills/flow/review-panel.md`).

**A plain-language session instruction overrides `DEFAULT_MODEL` for this run only** — "use opus for
the panel", "implement on haiku" — per **Model policy** (`skills/flow-contracts/model-policy.md`).
Record the instruction with the dispatch it changes; an override nobody wrote down is
indistinguishable from a mistake. The override is **never** written back to the settings store —
`/flow-settings` is the only command that changes a global default, per that command's own
guardrails.

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
  command** (`skills/flow-contracts/pipeline.md`): the change is archived. Proceed only on an
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
`<agents repo>/scripts/check-guard-symlinks.sh`'s rule 2 derives — is present in
`<skill-dir>/scripts/`. A complete set prints nothing; any absence prints that section's block once,
and the run continues under each guard's own hand-run fallback.

`check-unfinished-work.sh` and `check-task-commit-fields.sh` also require
`<agents repo>/scripts/lib/change-plan.sh` as a `<agents repo>/scripts/lib/` sibling — the same
sibling-dependency rule `<agents repo>/scripts/check-guard-symlinks.sh`'s rule 2 already applies to
every other guard above.

**The `<change>` argument to every mark below is always a resolved change name.** On a creating run
the name does not exist until **A. Resolve the change and write `STARTED`**
(`skills/flow/brainstorm.md`) produces it — defer `flow.state-gate`-equivalent bookkeeping into that
section, per **The `<change>` argument is always a
resolved change name** (`skills/flow-contracts/pipeline.md`). This router reads state above using
a guess or the best available name, which is legal for a read; it is never legal for a mark.

**Generate this run's session token once, right here, before the first mark any phase file below
makes — a short, unique literal string — and reuse that exact same value at every `stage begin` this
run makes, including inside every phase file it dispatches into.** One run, one token, never a fresh
one per mark or per phase file.

## Guardrails

- Never ask a planning-effort, model, or review-panel-roster question on a creating run. The
  roster is the recorded decision's, never asked — the settings store's on a `micro` decision; see **Model resolution** above and
  **Review panel** (`skills/flow/review-panel.md`).
- Never publish a proposal artifact. `artifactUrl` is written
  `null` and stays `null` for the life of the change.
- Never skip brainstorming's design gate, or leave `tasks.md` a thin scaffold.
- Never add a slot beyond the resolved roster automatically, by diff size, touched area, or any
  other trigger — only an explicit operator instruction adds one, for that run only, checked at the
  start of the panel stage and at every fix round. The one automatic change to the roster is a
  reduction — `check-panel-docs-only.sh`'s docs-only verdict dispatches `primary` alone — and it
  only ever removes, from the decision's roster as from any other; see **The docs-only reduction** (`skills/flow/review-panel.md`).
- Never run more than three implementer dispatches in flight at once, in any wave — a fourth or
  later ready group queues in plan order and launches only as an in-flight one is picked; see the
  Waves paragraph of **4. Execute (SDD + TDD)** (`skills/flow/implement.md`).
- Never dispatch review-panel roles as separate parallel `Agent` calls. A round is at most two
  dispatches, each one `Agent` call carrying one to three roles as its own `PASS <id>` sections —
  see **Bundled dispatch** (`skills/flow/review-panel.md`). Before dispatching any panel round,
  re-check the decision's `panel.dispatches`/`panel.grouping`.
- Never hand off with an open finding of any severity, or a stale clean result — no preset or
  fixed slot count moves this bar — stale as **Panel re-runs** (`skills/flow/review-panel.md`)
  defines it. A deferred Minor is not open.
- Never commit `<project>/spectre/changes/` in a task or fixup commit. Never push, merge,
  or open a PR outside the integrate/archive branches' own routes.
- Never advance the state past what the phase in force is entitled to write — a fix never moves
  the state; an implementation run only ever writes `IN_PROGRESS`; only run 2 of the archive branch
  writes `FINISHED`.
- `implement.md` sections 1, 2 and 4, <!-- refs-guard:allow -->
  `review-panel.md` and `verify-and-handoff.md` run in the parent session directly — the parent
  orchestrates every guard, gather, dispatch, mark and report itself and prints its own handoff,
  per **The parent orchestrates directly** and **Inline — the parent implements**
  (`skills/flow/implement.md`).
