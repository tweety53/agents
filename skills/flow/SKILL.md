---
name: flow
description: Single-command pipeline — brainstorm, implement behind the review panel resolved from the settings store, then integrate and archive across the same three-state pipeline, pausing only at the human gates. Re-run to resume, fix, or integrate. Use for /flow.
allowed-tools: Bash(spectre:*), Bash(flow:*)
license: MIT
---

Drive the three-state pipeline (`STARTED` → `IN_PROGRESS` → `FINISHED`) end to end in one command.
`/flow` replaces `/myflow-start`, `/myflow-do`, `/myflow-finish` and `/myflow-fast` — one command,
one state file, the same three states, content reorganized by topic rather than by the old
start/do/finish boundary. This file is the router: it
resolves state and dispatches into the topic file that owns the phase in force. Nothing here
duplicates that file's own content.

**Announce at start:** "Using flow for change `<name>`."

Immediately after that line, print these two commands for the operator to paste, per **Handoff
output** (`skills/flow-contracts/pipeline.md`) — that section fixes the colour and records why
they are printed rather than invoked:

```text
/rename <change-name>
/color cyan
```

**Load `skills/flow-contracts/pipeline.md` first** — canonical for the three states, the
transition table's shape, stage-mark mechanics, the guard-presence check, guard resolution, the
handoff shape and change-name resolution. Its **State transitions** table is `/flow`'s contract;
**Stage keys** below names which phase file marks each key.

**Then register this run's steps** with the harness's task-list mechanism, before any work begins,
and keep each entry's status current as the run proceeds, per **Progress visibility**
(`skills/flow-contracts/pipeline.md`). One entry per brainstorming checklist item and artifact on
the planning branch, one entry per `tasks.md` item on the implementation branch, one entry per step
of whichever finish run is executing on the integrate/archive branch — the same granularity
`/myflow-fast` used for the branch it is running, since `/flow` runs the same branches.

**No flags.** The only argument is the optional change name/description on a creating or resuming
run, or fix instructions at `IN_PROGRESS`; report anything else rather than ignoring it.

## Stage keys

Every stage `/flow` marks uses a `flow.*` key.

The full key list, in the order each phase file marks them:

| Phase file | Keys |
|------------|------|
| `skills/flow/brainstorm.md` | `flow.kickoff`, `flow.brainstorm`, `flow.design-approval`, `flow.create-artifacts`, `flow.writing-plans` |
| `skills/flow/implement.md` | `flow.load-context`, `flow.isolate-workspace`, `flow.document-fix`, `flow.sdd-tdd` |
| `skills/flow/review-panel.md` | `flow.review-panel` |
| `skills/flow/verify-and-handoff.md` | `flow.verify`, `flow.visual-verify`, `flow.stage-diff`, `flow.run-instructions`, `flow.write-in-progress` |
| `skills/flow/integrate.md` | `flow.preflight`, `flow.unfinished-work-gate`, `flow.landing-question`, `flow.preserve-sessions`, `flow.commit-two`, `flow.landing-routes` |
| `skills/flow/archive.md` | `flow.verify-merge`, `flow.sync-archive`, `flow.commit-archive`, `flow.cleanup`, `flow.verify-cleanup`, `flow.write-finished`, `flow.self-review`, `flow.push-archive` |

## Model resolution

**Resolve this once, near the top of every run, before any dispatch below reads it:**

```bash
MAIN_CHECKOUT="${MAIN_CHECKOUT:-$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd -P)}"
SETTINGS_JSON="$(flow settings get)"
DEFAULT_MODEL="$(printf '%s' "$SETTINGS_JSON" | jq -r '.defaultModel')"
REVIEWERS="$(printf '%s' "$SETTINGS_JSON" | jq -r '.reviewers[]')"
SELF_REVIEW_MODEL="$(printf '%s' "$SETTINGS_JSON" | jq -r '.selfReviewModel // empty')"
PROJECT_SRM="$(project-get.sh "$MAIN_CHECKOUT" 'self review model' 2>/dev/null | tr -d '`' | xargs)"
if [ -n "$PROJECT_SRM" ]; then
  if flow settings models | grep -qx -- "$PROJECT_SRM"; then SELF_REVIEW_MODEL="$PROJECT_SRM"
  else echo "⚠ flow: .flow/project.md '## self review model' body '$PROJECT_SRM' is not a valid model — dropped" >&2; fi
fi
[ -z "$SELF_REVIEW_MODEL" ] && SELF_REVIEW_MODEL=fable
PLANNING_MODEL="$(printf '%s' "$SETTINGS_JSON" | jq -r '.planningModel // empty')"
PROJECT_PM="$(project-get.sh "$MAIN_CHECKOUT" 'planning model' 2>/dev/null | tr -d '`' | xargs)"
if [ -n "$PROJECT_PM" ]; then
  if flow settings models | grep -qx -- "$PROJECT_PM"; then PLANNING_MODEL="$PROJECT_PM"
  else echo "⚠ flow: .flow/project.md '## planning model' body '$PROJECT_PM' is not a valid model — dropped" >&2; fi
fi
[ -z "$PLANNING_MODEL" ] && PLANNING_MODEL=fable
VERIFY_MODEL=sonnet
```

A non-zero exit from `flow settings get` means the settings store could not be reached — there is
no per-change fallback file for this record. Report the CLI's stderr and fall back to the literal
`sonnet` (the store's own no-row default, per `<agents repo>/stats/internal/store/settings.go`'s `DefaultModel`),
naming that this is a fallback rather than a resolved value, and continue: settings unreachable is
never a reason to block implementation. `SELF_REVIEW_MODEL` falls back to the literal `fable` on
the same failure, naming that this too is a fallback rather than a resolved value.

**`REVIEWERS` resolves from the same call, into the roster the panel dispatches**
(`skills/flow/review-panel.md` is canonical for what dispatching it means):

| Store state | Resolved roster |
|-------------|-----------------|
| Reachable, list non-empty | exactly the list |
| Reachable, list empty | `primary` alone |
| Unreachable | `primary`, `principles`, `code-review-low` (`DefaultReviewers` in `<agents repo>/stats/internal/store/settings.go`), naming this a fallback rather than a resolved value — the same pattern as `DEFAULT_MODEL`'s |

An empty list can never reach this table from `/flow-settings`: `<agents repo>/stats/cmd/flow/settings.go`'s
`settings set` refuses an empty `-reviewers` as a caller mistake before any write reaches the
store. The empty-list row exists because this resolver must still define a value for a state the
store's schema permits, not because an operator can produce one.

**`SELF_REVIEW_MODEL` resolves independently of `DEFAULT_MODEL` and governs the archive-phase
self-review subagent dispatch** (`skills/flow/archive.md` step 9). `<project>/.flow/project.md`'s
`## self review model` key, when present and a valid `ValidModels` member, wins over the store's
`selfReviewModel` field; when both are empty, or the store is unreachable, `SELF_REVIEW_MODEL` falls
back to the literal `fable`, naming this a fallback exactly as `DEFAULT_MODEL`'s own `sonnet` literal
is. A plain-language session instruction overrides `SELF_REVIEW_MODEL` for that run only, recorded
with the dispatch it changes and never written back.

**`PLANNING_MODEL` resolves independently of `DEFAULT_MODEL` and governs three roles**: the
planner dispatch (`skills/flow/brainstorm.md`), `flow.document-fix`'s planner dispatch
(`skills/flow/implement.md`), and `/flow-research`'s research subagent
(`skills/flow-research/SKILL.md`). `<project>/.flow/project.md`'s `## planning model` key, when
present and a valid `ValidModels` member, wins over the store's `planningModel` field; when both
are empty, or the store is unreachable, `PLANNING_MODEL` falls back to the literal `fable`, naming
this a fallback exactly as `DEFAULT_MODEL`'s own `sonnet` literal is. A plain-language session
instruction overrides `PLANNING_MODEL` for that run only, recorded with the dispatch it changes and
never written back to the settings store or the project key.

**`VERIFY_MODEL` governs the two verifier dispatches** — `flow.verify`'s and
`flow.visual-verify`'s (**Verify** and **Visual verification**, `skills/flow/verify-and-handoff.md`);
it is the fixed literal `sonnet`, read from neither the settings store nor
`<project>/.flow/project.md`; a plain-language session instruction does not override it; and it
never falls back, because it is never resolved — the point is a predictable model for mechanical
test and verification runs regardless of what `DEFAULT_MODEL` resolved to.

**`DEFAULT_MODEL` is the model for all three roles this run dispatches on** — the implementer
(`skills/flow/implement.md`), every panel slot that takes a model override, and the panel-fix
subagent (`skills/flow/review-panel.md`). Slots dispatched by `subagent_type` (Bugbot,
Security) take no override from this value; see **Review panel** (`skills/flow/review-panel.md`)
for why.

**A plain-language session instruction overrides `DEFAULT_MODEL` for this run only** — "use opus for
the panel", "implement on haiku" — the same override mechanism
`skills/flow-contracts/model-policy.md` already describes for the retired per-change fields,
reading from the settings store instead of a question round. Record the instruction with the
dispatch it changes; an override nobody wrote down is indistinguishable from a mistake. The override
is **never** written back to the settings store — `/flow-settings` is the only command that changes
a global default, per that command's own guardrails.

## Reading the state

```bash
flow state get <name-or-best-guess> -C <repo-root>
```

- **Exit 1**, or exit 0 with `"synthetic": true` — **no state**: a creating run. See
  **A. Resolve the change and write `STARTED`** (`skills/flow/brainstorm.md`); once `## Plan`
  returns, **Dispatch the conductor** (`skills/flow/implement.md`).
- **Exit 0, `"state": "STARTED"`** — a creating run interrupted before it reached `IN_PROGRESS`. See
  **Resuming at `STARTED`** (`skills/flow/brainstorm.md`).
- **Exit 0, `"state": "IN_PROGRESS"`, an argument present** — a fix run. See
  **3. Documenting a fix, before implementing it** and then **Dispatch the conductor**
  (`skills/flow/implement.md`).
- **Exit 0, `"state": "IN_PROGRESS"`, no argument** — an integrate run. See
  **Deciding which run this is** (`skills/flow/integrate.md`).
- **Exit 0, `"state": "FINISHED"`** — emit the wrong-state handoff from **Wrong state for this
  command** (`skills/flow-contracts/pipeline.md`): the change is archived. Proceed only on an
  explicit override.

**Check guard presence.** Per **Guard presence check** (`skills/flow-contracts/pipeline.md`),
confirm every guard `/flow` can invoke — the full list is the union carried by
`skills/flow/scripts/`: `check-base-moved.sh`, `check-cleanup-complete.sh`, `check-finish-preflight.sh`,
`check-panel-citation-trigger.sh`, `check-panel-diff-size.sh`, `check-panel-docs-only.sh`, `check-panel-findings-closed.sh`, `check-panel-reproducers.sh`, `check-plan-shape.sh`, `check-task-commit-fields.sh`,
`check-unfinished-work.sh`, `check-visual-trigger.sh`,
`check-visual-verification.sh`, `check-workspace-isolation.sh`,
`check-worktree-processes.sh`, `commit-split.sh`, `gather-dispatch-context.sh`, `gather-self-review-context.sh`,
`mutate-and-verify.sh`, `plan-dispatch-bundles.sh`, `prepare-archive-branch.sh`, `prepare-workspace.sh`,
`resolve-base-branch.sh`, `resolve-visual-screenshots.sh` and `run-reproducer.sh` — is present there. A complete set prints nothing;
any absence prints that section's block once, and the run continues under each guard's own hand-run
fallback.

`check-unfinished-work.sh` and `check-task-commit-fields.sh` also require
`<agents repo>/scripts/lib/change-plan.sh` as a `<agents repo>/scripts/lib/` sibling — the same
sibling-dependency rule `<agents repo>/scripts/check-guard-symlinks.sh`'s rule 2 already applies to
every other guard above.

**The `<change>` argument to every mark below is always a resolved change name.** On a creating run
the name does not exist until **A. Resolve the change and write `STARTED`**
(`skills/flow/brainstorm.md`) produces it — defer `flow.state-gate`-equivalent bookkeeping into that
section exactly as `/myflow-fast` deferred `do.state-gate`, per **The `<change>` argument is always a
resolved change name** (`skills/flow-contracts/pipeline.md`). This router reads state above using
a guess or the best available name, which is legal for a read; it is never legal for a mark.

**Generate this run's session token once, right here, before the first mark any phase file below
makes — a short, unique literal string — and reuse that exact same value at every `stage begin` this
run makes, including inside every phase file it dispatches into.** One run, one token, never a fresh
one per mark or per phase file.

## Guardrails

- **Never** ask a planning-effort, model, or review-panel-roster question on a creating run —
  `ask-options-removed`. The roster is resolved from the settings store, never asked; see
  **Model resolution** above and **Review panel** (`skills/flow/review-panel.md`).
- **Never** publish a proposal artifact — `publish-proposal-removed`. `artifactUrl` is written
  `null` and stays `null` for the life of the change.
- **Never** skip brainstorming's design gate, or leave `tasks.md` a thin scaffold — the removed
  stages are the options question round and the artifact publish, not the workflow steps
  themselves.
- **Never** add a slot beyond the resolved roster automatically, by diff size, touched area, or any
  other trigger — only an explicit operator instruction adds one, for that run only, checked at the
  start of the panel stage and at every fix round. The one automatic change to the roster is a
  reduction — `check-panel-docs-only.sh`'s docs-only verdict dispatches `primary` alone — and it
  only ever removes; see **The docs-only reduction** (`skills/flow/review-panel.md`).
- **Never** hand off with an open finding of any severity, or a stale clean result — no preset or
  fixed slot count moves this bar — stale as **Panel re-runs** (`skills/flow/review-panel.md`)
  defines it.
- **Never** commit `<project>/spectre/changes/` or `<project>/docs/superpowers/` in a task or fixup
  commit. **Never** push, merge, or open a PR outside the integrate/archive branches' own routes.
- **Never** advance the state past what the phase in force is entitled to write — a fix never moves
  the state; an implementation run only ever writes `IN_PROGRESS`; only run 2 of the archive branch
  writes `FINISHED`.
- **No flags.** The only argument is the optional change name/description, or fix instructions at
  `IN_PROGRESS`; report anything else rather than ignoring it.
- **Never** run `implement.md` sections 1, 2 or 4, <!-- refs-guard:allow -->
  `review-panel.md` or `verify-and-handoff.md` in the parent session — the conductor runs them; the
  parent dispatches it, relays its questions and prints its handoff.
