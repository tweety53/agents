---
name: flow-fast
description: Reduced-ceremony /flow variant — direct-write brainstorm with no design gate, inline TDD implementation with targeted-only tests/lint, a fixed primary+simple-reviewer panel, seven guards, and a one-run merge-and-push finish. Same state record and flow.* stage keys as /flow, so a change can move between the two commands. Use for /flow-fast.
allowed-tools: Bash(spectre:*), Bash(flow:*)
license: MIT
---

Drive the same three-state pipeline (`STARTED` → `IN_PROGRESS` → `FINISHED`) `/flow` drives, with
every choice `/flow` makes dynamically fixed instead: inline brainstorm, inline implementation, a
constant two-slot review panel, and a smaller guard set. `/flow-fast` is not a lighter copy of
`/flow`'s phase files — it restates nothing from `skills/flow/*.md`, sharing
`skills/flow-contracts/*` as canonical, the reviewer prompts under `skills/flow/`
(`simple-reviewer-prompt.md`, `principles-reviewer-prompt.md`, `engineering-principles.md`) and
`skills/flow/scripts/` unchanged. One change may move between `/flow` and `/flow-fast` at any
point in its life — both write the same state-file shape and the same `flow.*` stage keys, marked
`-command '/flow-fast'` here instead of `-command '/flow'`.

**Announce at start:** "Using flow-fast for change `<name>`." `/flow-fast` prints no
`/rename`/`/color` lines — it never had them, unlike `/flow`'s own deprecated pair (see
`skills/flow-contracts/pipeline.md`'s **Handoff output**).

**Load `skills/flow-contracts/pipeline.md` first** — canonical for the three states, the
transition table's shape, stage-mark mechanics, the guard-presence check, guard resolution, the
handoff shape and change-name resolution. Its **State transitions** table governs `/flow-fast`
exactly as it governs `/flow`; **Stage keys** below names which phase file marks each key.

**Then register this run's steps** with the harness's task-list mechanism, before any work begins,
per **Progress visibility** (`skills/flow-contracts/pipeline.md`) — one entry per step of whichever
phase file is running, at that phase file's own granularity. `/flow-fast` dispatches no planner
and no conductor, so nothing here needs the coarser per-stage granularity `/flow`'s conductor
dispatch forces: every step registers.

**No flags.** The only argument is the optional change name/description on a creating or resuming
run, or fix instructions at `IN_PROGRESS`; report anything else rather than ignoring it.

## Stage keys

Every stage `/flow-fast` marks uses the same `flow.*` keys `/flow` uses — never a `flow-fast.*`
prefix, since a change moving between the two commands must read one consistent stage vocabulary.
`/flow-fast` marks a strict subset — it never marks `flow.visual-verify`, `flow.verify-cleanup` or
`flow.self-review`.

| Phase file | Keys |
|------------|------|
| `skills/flow-fast/brainstorm.md` | `flow.kickoff`, `flow.brainstorm`, `flow.create-artifacts`, `flow.writing-plans`, `flow.decide` |
| `skills/flow-fast/implement.md` | `flow.load-context`, `flow.isolate-workspace`, `flow.document-fix`, `flow.sdd-tdd` |
| `skills/flow-fast/review.md` | `flow.review-panel`, `flow.verify`, `flow.stage-diff`, `flow.run-instructions`, `flow.write-in-progress` |
| `skills/flow-fast/finish.md` | `flow.preflight`, `flow.unfinished-work-gate`, `flow.landing-question`, `flow.preserve-sessions`, `flow.commit-two`, `flow.landing-routes`, `flow.verify-merge`, `flow.sync-archive`, `flow.commit-archive`, `flow.cleanup`, `flow.write-finished`, `flow.push-archive` |

`flow.design-approval` is never marked — `/flow-fast` runs no design-approval gate at all
(`skills/flow-fast/brainstorm.md`).

## Model resolution

**Resolve this once, near the top of every run, before any dispatch below reads it:**

```bash
MAIN_CHECKOUT="${MAIN_CHECKOUT:-$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd -P)}"
SETTINGS_JSON="$(flow settings get)"
DEFAULT_MODEL="$(printf '%s' "$SETTINGS_JSON" | jq -r '.defaultModel')"
```

**This is the whole of `/flow-fast`'s model resolution.** Unlike `/flow`'s own **Model resolution**
(`skills/flow/SKILL.md`), `/flow-fast` resolves no `PLANNING_MODEL`, `SELF_REVIEW_MODEL` or
`VERIFY_MODEL` — there is no planner subagent, no self-review subagent, and no visual-verification
subagent for any of those to govern — and reads no `## execution mode`, `## implementer model` or
`## review panel` project toggle: execution is always inline, there is no implementer/fixer
subagent for a model to govern, and the review roster is always `primary` + `simple-reviewer`,
never resolved from the settings store or a project toggle. `DEFAULT_MODEL` is the model for the
one role `/flow-fast` dispatches on multiple slots of: the review panel's `primary` slot (the
`simple-reviewer` slot is always `haiku`, fixed, per `skills/flow-fast/review.md`).

A non-zero exit from `flow settings get` means the settings store could not be reached — report the
CLI's stderr and fall back to the literal `sonnet` (the store's own no-row default), naming that
this is a fallback rather than a resolved value, and continue: settings unreachable is never a
reason to block implementation, exactly as `/flow`'s own resolution block states.

**A plain-language session instruction overrides `DEFAULT_MODEL` for this run only** — recorded
with the dispatch it changes, never written back to the settings store.

## Reading the state

```bash
flow state get <name-or-best-guess> -C <repo-root>
```

Read exactly as `/flow`'s own **Reading the state** (`skills/flow/SKILL.md`) describes — the same
five outcomes (no state, `STARTED`, `IN_PROGRESS` with an argument, `IN_PROGRESS` bare,
`FINISHED`) — but dispatch into this skill's own phase files instead of `/flow`'s:

- **Exit 1**, or exit 0 with `"synthetic": true` — a creating run. See **A. Resolve the change and
  write `STARTED`** (`skills/flow-fast/brainstorm.md`); once the plan is written, **Dispatch
  implementation** (`skills/flow-fast/implement.md`).
- **Exit 0, `"state": "STARTED"`** — resume the creating run from wherever brainstorming stopped,
  reading `spectre list --json` and `tasks.md` exactly as `/flow`'s own resumption rule does.
- **Exit 0, `"state": "IN_PROGRESS"`, an argument present** — a fix run. See
  `skills/flow-fast/implement.md`'s `flow.document-fix`.
- **Exit 0, `"state": "IN_PROGRESS"`, no argument** — an integrate/archive run. See
  `skills/flow-fast/finish.md`.
- **Exit 0, `"state": "FINISHED"`** — emit the wrong-state handoff from **Wrong state for this
  command** (`skills/flow-contracts/pipeline.md`). Proceed only on an explicit override.

**Check guard presence.** Per **Guard presence check** (`skills/flow-contracts/pipeline.md`),
confirm every guard `/flow-fast` can invoke — exactly `check-unfinished-work.sh`,
`check-base-moved.sh`, `check-finish-preflight.sh`, `check-archive-scope.sh`,
`check-workspace-isolation.sh`, `check-worktree-processes.sh` and
`check-panel-findings-closed.sh` (**Guard set** below) — is present in `<skill-dir>/scripts/`,
resolved against `skills/flow-fast/`'s own directory per **Guard resolution**
(`skills/flow-contracts/pipeline.md`): `skills/flow-fast/scripts/` carries its own symlink to each
of those guards (plus `prepare-workspace.sh`, `project-get.sh`, `prepare-archive-branch.sh`,
`resolve-base-branch.sh` and the shared `lib` sibling directory), pointing at the same underlying
files `skills/flow/scripts/`'s own symlinks point at —
never a second copy. A complete set prints nothing; any absence prints that section's block once,
naming `skills/flow-fast/scripts/` as the directory searched.

**`check-unfinished-work.sh` also requires `<agents repo>/scripts/lib/change-plan.sh`** as a
sibling, exactly as `/flow`'s own guard-presence check states.

**The `<change>` argument to every mark below is always a resolved change name**, per **The
`<change>` argument is always a resolved change name** (`skills/flow-contracts/pipeline.md`) — on a
creating run the name does not exist until `skills/flow-fast/brainstorm.md`'s section A produces it.

**Generate this run's session token once, right here, before the first mark any phase file below
makes**, and reuse that exact value at every later `stage begin` this run makes.

## Guard set

`/flow-fast` presence-checks and runs exactly seven guards, all resolved against
`skills/flow/scripts/`: `check-unfinished-work.sh`, `check-base-moved.sh`,
`check-finish-preflight.sh`, `check-archive-scope.sh`, `check-workspace-isolation.sh` (run
internally by `prepare-workspace.sh`, never invoked directly), `check-worktree-processes.sh` and
`check-panel-findings-closed.sh`. `check-cleanup-complete.sh` is **not** in this set — no
verify-cleanup pass runs for `/flow-fast`, whose one-run finish has nothing for it to verify.
Every other guard `skills/flow/scripts/` carries —
`check-panel-citation-trigger.sh`, `check-panel-diff-size.sh`, `check-panel-docs-only.sh`,
`check-panel-fix-single-dispatch.sh`, `check-panel-reproducers.sh`, `check-plan-shape.sh`,
`plan-class.sh`, `check-spec-reach.sh`, `check-task-commit-fields.sh`, `check-visual-trigger.sh`,
`check-visual-verification.sh`, `commit-split.sh`, `gather-dispatch-context.sh`,
`gather-self-review-context.sh`, `mutate-and-verify.sh`, `plan-dispatch-bundles.sh`,
`resolve-visual-screenshots.sh` and `run-reproducer.sh` — is dropped for `/flow-fast` outright,
never hand-run: no phase file below cites any of them.

## Guardrails

- **Never** dispatch a planner or conductor subagent — every stage runs in the parent session.
- **Never** ask a planning-effort, model, or review-panel-roster question — nothing here is
  dynamic; there is no toggle, no roll, and no settings-store roster to resolve for `/flow-fast`.
- **Never** run a design-approval gate, visual verification, self-review, or the verify-cleanup
  pass — none of the four exists for this command.
- **Never** widen the review roster beyond `primary` + `simple-reviewer`, by diff size, touched
  area, or any other trigger — not even an explicit operator instruction; a run that needs a wider
  panel is a `/flow` run, not a `/flow-fast` one.
- **Never** fix a Minor finding — every Minor is recorded and deferred, with no per-finding
  "trivially easy" exception, unlike `/flow`'s own default.
- **Never** run a full `## test` / `## lint` pass automatically — only the operator's own
  instruction text triggers one, never a stage boundary or the handoff.
- **Never** commit `<project>/spectre/changes/` or `<project>/docs/superpowers/` in a task or
  fixup commit — only in the planning-artifacts commit `skills/flow-fast/finish.md` makes.
- **Merge-and-push is the only landing route.** Never ask the landing question, never read
  `<project>/.flow/project.md`'s `## default landing route`, never push the change branch itself,
  and never open a pull request — a change that needs one is a `/flow` run.
- **Never** advance the state past what the phase in force is entitled to write — a fix never
  moves the state; brainstorm/implement/review only ever write `IN_PROGRESS`; only
  `skills/flow-fast/finish.md` completing writes `FINISHED`.
- **No flags.** The only argument is the optional change name/description, or fix instructions at
  `IN_PROGRESS`; report anything else rather than ignoring it.
