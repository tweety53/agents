# Review — fixed two-slot bundle, Critical/Major only, then handoff

Loaded by `skills/flow-fast/SKILL.md` once `skills/flow-fast/implement.md` hands off — every task
checked.

```bash
flow stage begin -command '/flow-fast' -stage flow.review-panel -harness <harness> -session-token ff-<literal-token> <name>
```

## The fixed roster

Always exactly two slots, bundled as **one** dispatch — never the settings-store reviewer list,
never the docs-only reduction, never widened by diff size, touched area, or an operator
instruction:

| Slot | Model | Briefed against | Job |
|---|---|---|---|
| `primary` | `DEFAULT_MODEL` | `final-review.diff`, `proposal.md`, `design.md`, each task's `**Files:**`/`**Tests:**`/`**Commit:**` fields | plan alignment only — never code quality (`skills/flow/review-panel.md`'s own **The roster** row for `primary`) |
| `simple-reviewer` | `haiku`, fixed | `final-review.diff` | high-confidence defects only, per `skills/flow/simple-reviewer-prompt.md` |

Write `<abs-worktree>/.superpowers/sdd/final-review.diff` before the first round, exactly as
`skills/flow/review-panel.md`'s own "Before writing `final-review.diff`" step does — cited, never
restated.

```bash
flow record dispatch begin -change <name> -role reviewer -slot primary+simple-reviewer \
  -model <DEFAULT_MODEL> -effort medium -session-token ff-<literal-token> -started-at <ts>
```

Dispatch both slots in one call, each carrying its own briefing, the **INDEPENDENT PASSES**,
TOOLS and MODEL HANDSHAKE paragraphs `skills/flow/review-panel.md`'s own dispatch
paragraphs use — cited verbatim from that file, never re-authored here. Findings are recorded one
`flow record finding` call per finding, `-slot primary` or `-slot simple-reviewer` naming which
role raised it, exactly as `skills/flow/review-panel.md`'s own record shape.

## Critical/Major fixed inline, Minor always deferred

**Critical and Major findings are fixed inline by the parent** — no panel-fix subagent, matching
`skills/flow-fast/implement.md`'s "inline" choice for the implementation phase itself. Fix one
finding at a time, smallest diff that resolves it, verified by that finding's own reproducer or the
narrowest test that exercises it.

**Every Minor is recorded and deferred — never fixed, with no per-finding judgment call.** This is
stricter than `/flow`'s own default (which fixes a Minor when it is trivially easy): `/flow-fast`
defers every one, unconditionally.

**After a fix round, re-run both slots on that round's delta** — the same per-slot-per-worktree
last-reviewed-sha mechanism `skills/flow/review-panel.md`'s **Panel re-runs** describes (cited, not
restated) — until no Critical or Major remains open:

```bash
check-panel-findings-closed.sh <worktree> <change>
```

Exit 0 proceeds below. Exit 1 means a finding still reads `open` — return to the fix loop for it.
Exit 2 stops the run.

```bash
flow stage end -command '/flow-fast' -stage flow.review-panel -outcome completed <name>
```

## Handoff

No `flow.visual-verify` — `/flow-fast` runs no visual verification.

```bash
flow stage begin -command '/flow-fast' -stage flow.verify -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.verify -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.stage-diff -harness <harness> -session-token ff-<literal-token> <name>
```

Stage the diff (`git -C <abs-worktree> add`, never `<project>/spectre/changes/` or
`<project>/docs/superpowers/` — those stage at integrate, per **Handoff output**'s git-boundaries
paragraph, `skills/flow-contracts/pipeline.md`).

```bash
flow stage end   -command '/flow-fast' -stage flow.stage-diff -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.run-instructions -harness <harness> -session-token ff-<literal-token> <name>
```

Print run instructions and the `IN_PROGRESS` block exactly as
`skills/flow-contracts/handoff-blocks.md` defines it — cited, never copied.

```bash
flow stage end   -command '/flow-fast' -stage flow.run-instructions -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.write-in-progress -harness <harness> -session-token ff-<literal-token> <name>
```

Write `IN_PROGRESS`, carrying `worktrees`, `jiraIssue` and every other field forward from
`STARTED`.

```bash
flow stage end -command '/flow-fast' -stage flow.write-in-progress -outcome completed <name>
```
