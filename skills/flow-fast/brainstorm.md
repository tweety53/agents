# Brainstorm and plan — inline, auto-pick, no design gate

Loaded by `skills/flow-fast/SKILL.md` on a creating run (no state) or a resuming run (`STARTED`).
Everything below runs in the parent session — **there is no planner subagent dispatch.**

## A. Resolve the change and write `STARTED`

Follow **Resolution (how `jiraIssue` is decided)** and **Change naming** in
`skills/flow-contracts/jira-integration.md` exactly — cited, never restated. **Transition the
issue to In Progress now**, before brainstorming, per that contract's **Transitions**; a failure is
one skipped-with-reason line and planning continues.

**Mark `flow.kickoff` now that the name is fixed, and write `STARTED` immediately:**

```bash
flow stage begin -command '/flow-fast' -stage flow.kickoff -harness <harness> -session-token ff-<literal-token> <name>
```

```json
{
  "state": "STARTED",
  "branch": null,
  "worktrees": {},
  "artifactUrl": null,
  "jiraIssue": "<resolved key, or null>",
  "planningEffort": null,
  "models": { "default": null },
  "prUrl": null,
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/flow-fast"
}
```

```bash
flow stage end -command '/flow-fast' -stage flow.kickoff -outcome completed <name>
```

### Resuming at `STARTED`

Exactly `skills/flow/brainstorm.md`'s own **Resuming at `STARTED`** rule: read `spectre list
--json`'s `done`/`total` and `tasks.md`'s own content to find where the run actually left off,
rather than assuming. State the resumption point plainly: "resuming `<name>` at `<point>`."

## B. Brainstorm — inline, auto-pick, no design gate

Seed from a staged research note first, exactly as `skills/flow/brainstorm-planner.md`
section B's "Seed from a staged research note, if one exists" describes — the same exact-filename
rule, the same multiple-match ask, the same "present the parsed structure" step. Carry a seeded
note's path forward to **C** for deletion. A seeded `<project>/docs/superpowers/research/<stem>/tasks.md` beside the note is taken in
**D** in place of writing-plans; a seeded `<project>/docs/superpowers/research/<stem>/decision.json` is ignored — `/flow-fast`'s
decision is fixed — and both go with the note in **C**.

Run `superpowers:brainstorming`'s checklist (items 1–8) **inline, in this session** — no planner
subagent. **At every options round, auto-pick the recommended option** rather than asking the
operator. Ask the operator only when:

- no option in the round is recommended (a genuine open-ended choice with no default), or
- the request cannot proceed at all without an answer (a missing target, an ambiguous scope that
  would ship something wrong if guessed).

**These two are the whole test.** A round whose options carry a recommended default is auto-picked
even when the decision is architectural, touches a shared/core file, or is hard to reverse — none
of those is a third ground to ask. Judgment about the *option itself* (which one is simplest,
safest, most consistent with the existing pattern) belongs in choosing what to recommend, before the
round is posed — never in deciding whether to pose it at all.

When asking is unavoidable, ask exactly as `superpowers:brainstorming` would — one question,
options named, a recommended default marked when one exists.

**There is no design-approval gate.** `/flow`'s merged convergence-and-approval confirm
(`skills/flow/brainstorm-planner.md`'s **Convergence**) is never run. Once the checklist's own
questions are exhausted (auto-picked or, rarely, asked), proceed directly into **C** — the
`IN_PROGRESS` staged-diff review is the only human gate this command has.

Save the design to `<project>/.worktrees/<name>/docs/superpowers/specs/YYYY-MM-DD-<name>-design.md`
if `superpowers:brainstorming` produces one, exactly as `/flow`'s own planner does; its content is
folded into `design.md` in **C** below and the loose file is not committed separately.

Mark `flow.brainstorm` around this section:

```bash
flow stage begin -command '/flow-fast' -stage flow.brainstorm -harness <harness> -session-token ff-<literal-token> <name>
# … the checklist, run inline, auto-picking …
flow stage end   -command '/flow-fast' -stage flow.brainstorm -outcome completed <name>
```

## C. Create the change and its artifacts

```bash
flow stage begin -command '/flow-fast' -stage flow.create-artifacts -harness <harness> -session-token ff-<literal-token> <name>
```

Then create the change worktree exactly as `skills/flow/brainstorm.md`'s
"The three returns" step 1–4 does — `check-worktree-location.sh`, `git check-ignore` /
`<project>/.git/info/exclude`, `git worktree add <project>/.worktrees/<name> -b spectre/<name>
<default-branch>`, `project-get.sh <worktree> "worktree setup"` and run its printed lines in the
foreground — cited, not restated. A worktree-setup command's non-zero exit ends the run naming the
command and its output, same as `/flow`'s own rule.

```bash
spectre new "<name>"
```

Write `proposal.md` (`## Why` / `## What changes`), `design.md` (`## Context` / `## Decisions` /
`## Open questions`) and `tasks.md` (checkbox scaffold) in exactly the shape
`skills/flow/brainstorm-planner.md` section C defines — the `## Decisions` / `## Open
questions` entry shape, the `ID:`/`Status:` fields, the immutability and supersede rules — cited,
never restated. A capability whose requirements this change alters still gets a task naming its
`<project>/spectre/specs/<capability>.md` edit in that task's own `**Files:**` field; the guard that would
otherwise check spec reach (`check-spec-reach.sh`) is not run, per `skills/flow-fast/SKILL.md`'s
**Guard set**, but the edit itself is not skipped.

**Delete the adopted staging note**, and its `<project>/docs/superpowers/research/<stem>/` directory when one exists, if one was
seeded in **B**, in this same commit.

```bash
flow stage end -command '/flow-fast' -stage flow.create-artifacts -outcome completed <name>
```

## D. Writing plans

Mark `flow.writing-plans`. When **B** carried a seeded `<project>/docs/superpowers/research/<stem>/tasks.md`, copy it to `tasks.md`,
fold in whatever **B** changed, and skip the invocation; otherwise invoke
**superpowers:writing-plans** to enrich `tasks.md` to plan
quality exactly as `skills/flow/brainstorm-planner.md` section D describes — the task shape
(`- [ ] <n>. <title>` with `  - [ ] **Step N: …**` children), the per-task verify-step rule
(targeted lint plus the build tool's own test selector, never the project's whole `## lint` /
`## test` list), the UI-test follow-on-task rule, plan-provenance tagging
(`skills/flow-contracts/plan-provenance.md`), the build-green tag and the mechanically-checkable
field family (`**Files:**`, `**Tests:**`, `**Regression:**`, `**Baseline:**`, `**Commit:**`, the
`**Squash-with:**` pairing for `Build: red`, the optional `**After:**`) and the two required header
lines (`**Execution:**`, `**Relocation:**`) — cited in full, never restated. Run
`check-plan-shape.sh` unconditionally and the project's configured plan-provenance/build-green
guards where declared; fix any hit.

**No `## execution mode` / `## implementer model` / `## review panel` roll runs here.** Write
`<abs-worktree>/.superpowers/sdd/decision.json` directly, with the fixed decision
`skills/flow-fast/SKILL.md`'s **Model resolution** and `design.md`'s `flow-fast-state-interop`
name: `execution: "inline"`, `implementer: "skipped — inline"`, `panel: {roster:
[{slot:"primary", model: <DEFAULT_MODEL>, effort:"medium"}, {slot:"simple-reviewer",
model:"haiku", effort:"medium"}], grouping:"static", dispatches: [["primary","simple-reviewer"]],
rerun:"delta", compact: true}`, `groups: null`, `overrides: {}` — no rolls, no class computation:
`/flow-fast` runs no classifier at all. Record it:

```bash
flow record decision -change <name> -session-token ff-<literal-token> -file <abs-worktree>/.superpowers/sdd/decision.json
```

so a `/flow` run resuming this change later reads a valid record (`design.md`'s
`flow-fast-state-interop`).

```bash
flow stage end -command '/flow-fast' -stage flow.writing-plans -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.decide -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.decide -outcome completed <name>
```

Continue directly into `skills/flow-fast/implement.md` — there is no gate here.

## Resume and fix runs

A fix run at `IN_PROGRESS` documents the fix in `proposal.md`/`tasks.md` (or a `<name>-fix-N`
sub-change) exactly as `/flow`'s own `flow.document-fix` does, then hands the appended plan to
`skills/flow-fast/implement.md`. The fixed decision above is re-recorded unchanged — there is
nothing for a fix run to re-roll, since `/flow-fast` never rolled anything to begin with.
