# KAN-182 — one change, two rows: the state-gate mark that names a guess

**Date:** 2026-08-15
**Change:** `kan-182-duplicate-change-row-from-best-guess-mark`

## The observation

The stats dashboard's "Every open change" table showed project `gymie-7c1f238a` twice:
`kan-175-more-ui-ux-fixes` at 18:06:23 and `kan-175` at 18:05:54, both `STARTED`, both stamped
`myflow stage begin (synthetic)`, both offering `/myflow-do` as the next command. One of them is a
change; the other is a name somebody guessed 29 seconds earlier.

## The mechanism, end to end

1. `/myflow-fast`'s **State gate** section runs before anything resolves the change name, because
   its whole job is to decide which branch the run takes. It therefore refers to the change as
   `<name-or-best-guess>` — both in the read (`myflow state get`) and in the marks
   (`myflow stage begin -stage do.state-gate`).
2. The read is harmless: `state get` on an unknown name exits 1 and writes nothing. That exit is
   precisely the signal the gate wants — "no state, this is a creating run".
3. The mark is not. `ApplyBeginStageMark` (`stats/internal/api/stages.go`) catches
   `store.ErrChangeNotFound` and bootstraps a change row — state `STARTED`, `updated_by =
   stages.SyntheticChangeUpdatedBy` — so the mark has something to attach to, then retries once.
4. So the guessed name becomes a row in `changes`, indistinguishable from a real one to every query
   except by `updated_by`. It is never archived, because no change directory bearing that name
   exists to archive.
5. Section A then resolves the Jira key into `<key>-<slug>`, and every subsequent mark carries the
   real name — which bootstraps a *second* row.

## Why the fix does not belong in the daemon

The bootstrap is kan-174's recorded decision, and `stats/internal/stages/synthetic.go` states its
reasoning in full: a mark is never dropped, and the resulting row is deliberately
indistinguishable from a real one so that "a defect worth seeing in the data" is visible rather
than being a special case a reader has to know to look for.

It did exactly that here. The row is not the bug; it is the bug's report. Teaching the daemon to
refuse the bootstrap for `do.state-gate`, or teaching the dashboard to filter synthetic rows out of
the open-change list, would each remove the report and leave the cause — a skill that names
something that is not a change — in place, and would leave it in place for every *other* mistyped
change name too, silently.

The cause is one placeholder in one section of one skill. `/myflow-do` and `/myflow-start` already
mark with a resolved `<name>` throughout; `/myflow-start` section A states the rule explicitly for
its own creating run: `begin` and `end` "fire together, back to back, only now — never before, since
`myflow stage begin`'s `<change>` argument requires a name that does not exist until this point."
`/myflow-fast` is the one command that violates a rule its own sibling already wrote down.

## The design

### 1. Defer the mark, not the gate

In `skills/myflow-fast/SKILL.md`'s **State gate**:

- The read stays as it is, on `<name-or-best-guess>`. It is the only step that can run before a
  name exists, and it creates nothing.
- The `do.state-gate` `begin`/`end` pair takes a resolved `<name>`.
- On a **creating run**, the pair fires back to back immediately after section A fixes the change
  name, before `start.resolve-change`'s own marks — the same shape, and for the same reason, as
  `/myflow-start` section A's creating-run rule.
- At **`IN_PROGRESS`**, the name is resolved before the gate in every path (an argument names the
  change, a bare invocation resolves it through **Change name resolution**, which enumerates names
  that already exist), so those marks stay exactly where they are.

What is lost: on a creating run, `do.state-gate`'s recorded duration collapses to roughly zero,
because the mark now brackets nothing. That is already true of `start.resolve-change` on a creating
run, for the identical reason, and a stage run that records a truthful zero is better than one that
records a duration by manufacturing a change row to hang it on.

### 2. Make the rule mechanical

`scripts/check-stage-mark-calls.sh` already enforces three rules over `stage begin` calls in skill
source: `-session-token` present and literal, `-harness` present and *not* literal. It gains a
fourth: the positional change argument may not be a placeholder that names a guess — the shape
`<...guess...>`, case-insensitively, inside angle brackets.

The check is deliberately narrow. It cannot know whether `<name>` at a given call site is resolved
by that point in the skill; it can know that a placeholder whose own text says "guess" is one that
is not. That is enough to fail today's `skills/myflow-fast/SKILL.md`, which is what makes this
change testable, and enough to stop the placeholder being reintroduced.

### 3. State the rule once

One paragraph under **Stage marks** (`skills/myflow-contracts/pipeline.md`), beside the
session-token and harness rules, saying that a mark's `<change>` argument is a resolved change name
and what a guessed one costs. That file is where the guard's other three rules are stated, so this
is the section a reader is already in when they need it.

### 4. Delete the stray row

One transaction against the shared development database — `myflow` on `localhost:5433`, explicitly
not an apply worktree's isolated `myflow_<id>` — removing `gymie-7c1f238a` / `kan-175`. There is no
`ON DELETE CASCADE` on `stage_runs.change_id` or `change_repos.change_id`, so the order is
`stage_runs`, `change_repos`, `changes`. `kan-175-more-ui-ux-fixes` is matched by exact name and is
not touched.

No `myflow state delete` command is added. One row needs removing, and the cause is being fixed in
the same change; a delete path into the store is surface area nobody has asked for twice.

## Verification

- `scripts/test-check-stage-mark-calls.sh` — new fixtures for the fourth check, written before the
  guard learns it, and a fixture proving a compliant call still passes.
- `scripts/check-stage-mark-calls.sh` over `skills/` — red before the skill edit, green after.
- The full `## lint` list from `.myflow/project.md`, `check-contract-budget.sh` included.
- `cd stats && go test ./... -race -count=1` — no Go source changes, run as a regression check.
- The dashboard, after the delete: one `kan-175-*` row.
