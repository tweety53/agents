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

Write `final-review.diff` per **Writing `final-review.diff`** (`skills/flow/panel-dispatch.md`) —
cited, never restated.

```bash
flow record dispatch begin -change <name> -role reviewer -slot primary+simple-reviewer \
  -model <DEFAULT_MODEL> -effort medium -session-token ff-<literal-token> -started-at <ts>
```

**One dispatch, one subagent — never two parallel calls.** The `-slot primary+simple-reviewer`
above is the tell: both roles run in the same turn, as two `## PASS <role>` sections (`PASS
primary`, `PASS simple-reviewer`), on the one model this dispatch record names — cited in full
under **Bundled dispatch** (`skills/flow/review-panel.md`). Each PASS carries its own briefing —
`primary` against `final-review.diff`/`proposal.md`/`design.md`/each task's fields, `simple-reviewer`
against `final-review.diff` alone — plus the **Dispatch paragraphs** (`skills/flow/panel-dispatch.md`)
— INDEPENDENT PASSES, TOOLS and MODEL HANDSHAKE — cited verbatim from that file, never re-authored
here. Findings are recorded one `flow record finding` call per finding, `-slot primary` or `-slot
simple-reviewer` naming which role raised it, exactly as `skills/flow/review-panel.md`'s own record
shape.

## Critical/Major fixed inline, Minor always deferred

**Critical and Major findings are fixed inline by the parent** — no panel-fix subagent, matching
`skills/flow-fast/implement.md`'s "inline" choice for the implementation phase itself. Fix one
finding at a time, smallest diff that resolves it, verified by that finding's own reproducer or the
narrowest test that exercises it.

**Every Minor is recorded and deferred — never fixed, with no per-finding judgment call.** This is
stricter than `/flow`'s own default (which fixes a Minor when it is trivially easy): `/flow-fast`
defers every one, unconditionally.

**After a fix round, re-run only the `simple-reviewer` slot on that round's delta** — `primary`'s
pass-1 plan-alignment verdict stands; a defect fix does not move plan alignment, so re-running
`primary` finds nothing new. The same per-slot-per-worktree last-reviewed-sha mechanism
`skills/flow/review-panel.md`'s **Panel re-runs** describes (cited, not restated) applies to the
`simple-reviewer` slot alone — until no Critical or Major remains open:

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
`<project>/docs/superpowers/` — those stage at finish, per **Handoff output**'s git-boundaries
paragraph, `skills/flow-contracts/pipeline.md`).

```bash
flow stage end   -command '/flow-fast' -stage flow.stage-diff -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.run-instructions -harness <harness> -session-token ff-<literal-token> <name>
```

Resolve run instructions exactly as **Resolve the run instructions**
(`skills/flow/verify-and-handoff.md`) does — cited, never restated: `## stop` when declared, then
`## run`.

```bash
flow stage end   -command '/flow-fast' -stage flow.run-instructions -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.write-in-progress -harness <harness> -session-token ff-<literal-token> <name>
```

Write `IN_PROGRESS`, carrying `worktrees`, `jiraIssue` and every other field forward from
`STARTED`.

```bash
flow stage end -command '/flow-fast' -stage flow.write-in-progress -outcome completed <name>
```

Print this block verbatim, filled in — carried here rather than citing
`skills/flow-contracts/handoff-blocks.md`, since `/flow-fast` has one fixed panel, one fixed guard
set and one landing route, so nothing in that file's `(run-only)` variance applies:

```text
## Implementation staged — review and test

**Change:** <name>
**Panel:** primary+simple-reviewer · fixed, compact, delta rerun (simple-reviewer only)
**Staged:** <completed>/<total> tasks · staged and uncommitted
**Records:** <all writes reached the store, or "N write(s) journalled — the store was unreachable">
**Deferred:** <count of deferred Minors>

Worktree:   <absolute worktree path>

Running:
  <url line>  # <from the start command's output>
  <stop command>

Review the diff, then run it:
  git -C <absolute worktree path> diff --cached
  open -na "IntelliJ IDEA" --args "<absolute worktree path>"

### Deferred minors
<one row per deferred Minor, `F<n> <location> — <note> — <reason>`, or `none` when there are none>

Next:
/flow-fast <name>
```
