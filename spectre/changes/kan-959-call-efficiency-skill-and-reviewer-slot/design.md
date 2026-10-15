## Context

KAN-959 is the KAN-924 audit's Prevention recommendation, "both": (A) a skill carrying the
measuring method and (B) a reviewer slot asking the two questions a diff can answer. The operator
added, at kickoff, that `test-sanity` and the new slot ride one extra dispatch on every run's panel,
beside the basic primary+principles, on `sonnet`/`low` or `opus`/`low` at 50/50, always that combo.

Constraints that shape it:

- `plan-class.sh` pins its existing output lines byte for byte (`plan_class_test.go`): the rolls
  `/flow`'s Decide records must not move, so the new roll gets its own line rather than joining
  `rolls:`.
- A pass-1 panel dispatch is `opus`, always, under **Model and effort**
  (`skills/flow/brainstorm-planner.md`) and the operator's Models rule; this change carves the
  extra dispatch out of both.
- A `micro` decision's `panel` is the string `default` and records no pair, so the extra
  dispatch's pair cannot live under `panel`.
- `render-slot-prompt.sh` resolves a slot's template as `<skill-dir>/<id>-reviewer-prompt.md`; an
  id outside its known set is refused.

## Decisions

### The extra dispatch is a top-level decision pair

**ID:** extra-dispatch-top-level-pair
**Status:** active
**Chosen:** a top-level `extra_dispatch` object `{slots, model, effort, reason}` in `decision.json`,
recorded on every class, `micro` included — one place, one shape, read the same way by a decided and
a `default` panel.
**Considered:** a third entry in `panel.dispatches` — rejected: `micro`'s `panel` is a string with no
`dispatches`, so the pair would need a second home for `micro`, and the `≤2 × ≤3` dispatch cap every
grouping rule states would need an exception; adding the two ids to the settings store's
`REVIEWERS` — rejected: that reaches only `micro`, carries no model roll, and the operator could
drop them.

### The roll is its own plan-class line

**ID:** extra-roll-own-line
**Status:** active
**Chosen:** `plan-class.sh` prints an eighth line, `extra: test-sanity+call-efficiency · roll <N> ·
<sonnet|opus> low`, `<N>` the first 8 hex digits of SHA-256(`<name>extra`) mod 100, `sonnet` when
`<N> < 50` — on every class, `micro` included.
**Considered:** appending `· extra N` to the `rolls:` line — rejected: that line is pinned byte for
byte and moving it would move every recorded roll's parse; a non-reproducible random pick at
dispatch time — rejected: the decision must record what ran, and a re-run of Decide must land the
same pair.

### The extra dispatch is never reduced

**ID:** extra-dispatch-always
**Status:** active
**Chosen:** pass 1 dispatches the extra bundle on every panel — decided or `default`, and under the
docs-only reduction, the late-fix reduction and the small append alike; `/flow-fast`'s `micro`,
which runs no panel today, runs the extra dispatch alone.
**Considered:** dropping it on a docs-only diff, where neither slot usually has anything to read —
rejected: the operator asked for "always this combo" on "all the runs"; each slot reports an empty
result in one cheap `low` pass when the diff gives it nothing.

### Fix-round re-runs use the existing rerun rules

**ID:** extra-rerun-unchanged
**Status:** active
**Chosen:** a role from the extra dispatch that raised a finding re-runs like any other role: in the
decided panel's `panel.rerun_dispatch` bundle, and on `opus` on a `default` panel.
**Considered:** re-running it on its own rolled pair — rejected: a second rerun shape for two roles,
with no reason the existing one does not fit.

### Sonnet on this one pass-1 dispatch

**ID:** extra-dispatch-sonnet-carve-out
**Status:** active
**Chosen:** **Model and effort** names the extra dispatch as the one pass-1 panel dispatch that may
run on `sonnet`, at `low`, by its roll; the store already accepts `sonnet` in any pair, so the
carve-out is prose plus a store check that `extra_dispatch.model` is a valid model. The operator's
Models rule in `~/.claude/CLAUDE.md` (and its `~/.zcode` twin, which `check-hand-notes-in-step.sh`
keeps in step) gains the same carve-out, edited by the parent outside this repository once the
change lands.
**Considered:** opus-low always — rejected by the operator at kickoff.

### test-sanity is promoted, the experimental mechanism stays

**ID:** test-sanity-promoted
**Status:** active
**Chosen:** `skills/flow/experimental/test-sanity.md` moves to
`skills/flow/test-sanity-reviewer-prompt.md`, slot id `test-sanity`; the experimental directory is
left empty and the experimental roll reads `none available` until a new candidate is added.
**Considered:** keeping it as an experimental candidate too — rejected: one slot on the same panel
twice.

### Where the new files live

**ID:** call-efficiency-file-homes
**Status:** active
**Chosen:** the skill at `skills/call-efficiency-audit/SKILL.md` (`/call-efficiency-audit`), a sibling of the flow
skills that `setup.sh` already installs; the slot template at
`skills/flow/call-efficiency-reviewer-prompt.md`, beside the other slot templates.
**Considered:** the method inside the slot template — rejected: the reviewer reads a diff and
cannot run a stack; the measuring method is a separate, operator-invoked job.

## Live check

The extra dispatch is exercised by this change's own review panel: `plan-class.sh` on this
change's `tasks.md` prints the `extra:` line, Decide records `extra_dispatch`, `flow record
decision` accepts it, and `flow decision render` prints its line. A change that did not work
looks like: no `extra:` line, a refused decision, or a render with no extra-dispatch line.

## Open questions
