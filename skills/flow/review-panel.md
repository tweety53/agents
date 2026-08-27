# Review panel

Loaded by `skills/flow/SKILL.md` immediately after `skills/flow/implement.md`'s `flow.sdd-tdd`
stage closes, on every implementation run — creating, resumed, or fix. Fixed at **3 required
slots**, always, per design.md's `review-panel-fixed-3` — this file carries no roster table and no
diff-size/touched-area trigger table; both are gone, not reduced.

```bash
myflow stage begin -command '/flow' -stage flow.review-panel -harness <harness> -session-token mf-<literal-token> <name>
```

**Rebuild the dispatch context bundle at the start of this stage too** — never reused from
`skills/flow/implement.md`'s run. Overwrite the same path:

```bash
mkdir -p <worktree>/.superpowers/sdd
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  > <worktree>/.superpowers/sdd/dispatch-context.md
```

## The three required slots

Every `/flow` run dispatches exactly these three, unconditionally — no preset, no trigger, nothing
to select:

| # | Slot | Model | How to spawn |
|---|------|-------|---------------|
| 0 | **Primary** — plan alignment + code quality | `DEFAULT_MODEL` | **superpowers:requesting-code-review** with `final-review.diff` + the plan/spec constraints |
| 2 | **Principles** | `DEFAULT_MODEL` | general-purpose + `principles-reviewer-prompt.md`; all three principle groups always apply, all three principle groups are always covered <!-- refs-guard:allow --> |
| 3 | **Code review (low)** | `DEFAULT_MODEL` | general-purpose reviewer briefed for high-confidence defects only, against `final-review.diff` |

`DEFAULT_MODEL` is `skills/flow/SKILL.md`'s **Model resolution** value for this run — the settings
store's `defaultModel`, or this run's session-instruction override. There is no parent-model
inheritance and no economy tier.

Numbering follows the retired roster table's own slot numbers (0, 2, 3) rather than renumbering —
slots 1 and 4 below keep their old numbers too, and a gap in the sequence is the visible trace of
what the fixed-3 decision removed (the old `standard`/`full` roster's Bugbot-as-required-slot-1
option, and `full`'s auto-triggered Security, Adversarial and two extra principle slots).

## The two on-demand slots

| # | Slot | Included when |
|---|------|----------------|
| 1 | **Bugbot** — defect hunt | the operator explicitly names it |
| 4 | **Security** | the operator explicitly names it |

**Neither is ever included by a diff-size, touched-area, or any other automatic trigger.** This is
design.md's `review-panel-fixed-3` in full: "the on/off decision is entirely explicit, not
heuristic-driven." The retired roster's Adversarial slot and its two extra principle slots are cut
entirely, not folded into this on-demand pair — `/flow`'s panel has exactly five possible slots
(0, 1, 2, 3, 4), never six or seven.

**Check for an explicit instruction at two points**: at the start of this stage (has the operator,
in this run's own argument or in the session before this stage, named Bugbot and/or Security?), and
again at the start of every fix round below — an operator may ask for one mid-run, after seeing pass
1's result, and that request adds it starting from the round it was made, never retroactively to a
pass already closed. Record which on-demand slots are included and why (the operator's own words),
and record explicitly when neither is included: "Bugbot: not requested. Security: not requested." —
a documentation-, prompt-, or test-only diff with no operator request runs the three required slots
alone, and that is a correct outcome, said so explicitly, never a silently skipped review.

Spawn shape, unchanged from the retired roster's own table: Bugbot as `subagent_type: bugbot`,
`Diff: uncommitted changes`, `Full Repository Path: <worktree>`, plus the mutation-testing brief
below; Security as `subagent_type: security-review`, same shape. Both are dispatched by
`subagent_type` and carry their own agent definitions — pass them **no** model override, and record
`unknown (agent-defined)` as their dispatch row's `-model`.

**Before writing `final-review.diff`**, run

```bash
check-panel-diff-size.sh <worktree> <merge-base>
```

Exit 0 proceeds. Exit 1 puts the choice to the operator, shape per Operator prompts
(`skills/myflow-contracts/operator-prompts.md`):

> **The panel diff measured `<count>`, over the `<cap>` cap. How should this proceed?**
> - **Proceed with the panel anyway** *(default, recommended)*
> - **Stop and split the change** — ends the run at `IN_PROGRESS` with the implementation committed
>   on the branch

Exit 2 stops the run. Record the measured count, the cap in force, and the operator's answer where
one was given in `<abs-worktree>/.superpowers/sdd/final-review-panel.md` on **every** run, including
exit-0 runs.

Write `<abs-worktree>/.superpowers/sdd/final-review.diff` from `git diff <merge-base>` (staged and
unstaged), then dispatch **separate** review subagents — one per included slot, in **every**
affected worktree. Never merge two slots into one prompt.

**Every slot's dispatch is recorded**, the same pair section 4 of `skills/flow/implement.md`
records for an implementer:

```bash
myflow record dispatch begin -change <name> -role reviewer -slot <slot> -model <m> \
  -agent-id <id> -diff-base <sha> -key panel-<round>-<slot> \
  -session-token mf-<literal-token> -started-at <ts>
myflow record dispatch end -change <name> -key panel-<round>-<slot> \
  -session-token mf-<literal-token> -outcome completed -ended-at <ts> -agent-id <id>
```

`-slot` is `Primary`, `Bugbot`, `Principles`, `Code review (low)`, or `Security`. `-role` is
`reviewer` for every one; `-task` is omitted. `-diff-base <sha>` is the sha a delta-slot's delta
starts from, passed on a slot dispatched against a delta and on no other. `-model` is `DEFAULT_MODEL`
(or this run's override) for every slot except Bugbot and Security, which record `unknown
(agent-defined)`.

**On Claude Code, `-agent-id` is the identifier an asynchronous agent launch returns in the parent's
own tool result, at launch.** Never invent one.

**Record a slot's dispatch before recording that slot's findings**, and carry the seq the command
printed — `recorded: dispatch <seq>` — into each of that slot's `myflow record finding` calls as
`-dispatch-seq <seq>`.

**Every slot must supply, per finding, a reproducer**: a runnable command that demonstrates the
defect, or the literal exemption form `none — <reason>`. Carry this requirement on every slot's
dispatch prompt.

**Every slot's dispatch prompt also carries the CONTEXT BUNDLE paragraph** — the same one
`skills/flow/implement.md`'s implementer dispatch carries.

### No forking, and a wall-clock ceiling on every slot

**No panel slot SHALL be dispatched onto a skill or agent that forks its own background agent.**
The repair is to dispatch that slot on a shape that reports back to the dispatcher directly. **Never
drop the slot.**

**Every panel slot SHALL carry a 15-minute wall-clock ceiling from its dispatch.** The dispatcher
tracks each in-flight slot's elapsed time itself rather than blocking indefinitely on a completion
notification.

On a breach, in order: stop the slot; close its dispatch row (`-outcome timed-out`); record the
breach in the panel record, naming the slot and its elapsed time; re-dispatch that one slot once.

A second breach of the same slot is put to the operator, shape per Operator prompts
(`skills/myflow-contracts/operator-prompts.md`):

> **Slot `<slot>` breached the wall-clock ceiling a second time. How should this proceed?**
> - **Re-dispatch it again**
> - **Proceed without the slot**
> - **Stop the run** *(default, recommended)* — ends at `IN_PROGRESS` with the implementation
>   committed on the branch

A timed-out slot raises no finding, consumes no fix round, and is not a clean result for the final
pass.

### Code review (low)

Slot 3 is unconditionally a `general-purpose` subagent on `DEFAULT_MODEL`, briefed to report
high-confidence defects only against `final-review.diff`. It invokes no skill. Its findings are
ordinary `F<n>` rows, exactly like every other slot's.

### Bugbot's mutation-testing brief

Wherever the panel dispatches Bugbot, its dispatch prompt carries a mutation-testing brief: for
each behaviour the diff changes, mutate it — flip a condition, drop a guard, move a boundary, remove
a branch — and establish whether an existing test fails. A mutation no test catches is a
**surviving mutant**, an ordinary finding that blocks the handoff exactly as any other, unless the
operator withdraws it with a reason.

Slot 2 is the panel's only mandatory judgment check on *how* the code is built. It reads
`engineering-principles.md` — never a pasted copy — and owns the project's **hard invariants** from
its standards files.

**Resolve `[PRINCIPLES_PATH]` before dispatching the principles slot.** It is the **absolute** path
of `engineering-principles.md` **beside this file** — `skills/flow/`, always. Confirm the file
exists before spawning; if it does not, stop and report rather than dispatching a blind reviewer.

**Resolve `[STANDARDS_PATHS]` before dispatching slot 2**, from the `## standards` entries in
`<project>/.myflow/project.md`, per **Project configuration**
(`skills/myflow-contracts/project-configuration.md`). Pass an **empty** value when none resolve.
Record which standards files were passed, or that none resolved.

## Recording findings, and the record's format

**Every finding is a row in the store. The panel record is rendered from those rows.** As each slot
raises a finding, record it — one call, as it is raised:

```bash
myflow record finding -change <name> -ref F<n> -round <r> -slot <slot> -severity <sev> \
  -location <file:line> -status open -reproducer <command | none — reason> \
  -dispatch-seq <seq> -note <the finding>
```

`-round` is `0` for the initial panel and `1..n` for a fix round. `-ref` is unique within the change;
the store's own constraint enforces it. **A fix round updates the finding it resolved rather than
appending a second row:**

```bash
myflow record status -change <name> -ref F<n> -status fixed
```

**`-status` carries the whole status text the marker line shows** — a withdrawal passes its reason
with it: `-status 'withdrawn <the operator's reason>'`.

**Render the record when the panel closes** — every slot's result clean, no finding open:

```bash
myflow record render -change <name> -kind panel -repo <abs-worktree>
```

**A panel that raised nothing still renders**, declaring `findings-total: 0`. There is no
skip-the-render shortcut.

The record carries a findings table, one row per finding:

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Bugbot | Minor | `src/Foo.kt:42` | replaced the silent catch |

and, below it, the marker block — one line per row, plus the count:

```
findings-total: 1
finding-status: F1 fixed
```

The marker format is never quoted inside the record itself. The renderer neutralises any marker
label a finding's note or location happens to carry, on the way out only.

**The reproducer each finding's slot supplied gets a marker block of its own**, separate from the
`finding-status:` block above:

```
reproducers-total: 1
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh
```

A finding recorded with no reproducer renders the `none — <reason>` exemption form.

`check-unfinished-work.sh` reads **only the marker block** — never the table's header, column order,
cell boundaries, or where it starts or stops. Run it and read its own output for the reject reason:

- `<status>` is **exactly** `open`, `fixed` or `withdrawn`, compared byte for byte.
- `withdrawn` **carries its reason on the same line**.
- Each `F<n>` names **one** finding.
- The marker lines sit on **consecutive lines**.
- `findings-total: <n>` appears **exactly once** and equals the number of marker lines.

**The table carries no status column, on purpose.** To read a finding's state, look up its `F<n>`
in the marker block.

**A `withdrawn` marker's reason is checked for being there at all.** Fix subagents record `fixed`
when they fixed it, and leave `open` when they did not.

## Panel re-runs

**Pass 1 always runs the three required slots plus every on-demand slot the operator named at this
stage's start.** Only re-runs after a fix are scoped. Record `FIX_BASE=<task-sha>` — the task's
commit as it stood before this fix round — then, once the fix is folded into that commit via `git
commit --fixup=<task-sha>` and `git rebase --autosquash`, write
`<abs-worktree>/.superpowers/sdd/fix-round-N.diff` from `git diff "$FIX_BASE"..<task-sha>`.

When a review finding requires a code change to a task that is already committed, commit the fix as
`git commit --fixup=<task-sha>`, where `<task-sha>` is the **original** task commit. Immediately
`git rebase --autosquash` to fold it in.

| Mode | Who re-runs | Diff they get |
|------|-------------|----------------|
| **Targeted** (default) | Slot 0 (always, as integration check) + every agent that raised a finding | `fix-round-N.diff` |
| **Full** (escalation) | Every **required** slot, plus every on-demand slot already dispatched in an earlier pass of this run | Slot 0 the rewritten `final-review.diff`; every other diff-reading slot its own delta, below |

**A delta is `git diff <the HEAD sha that slot last reviewed> HEAD`**, written to
`<abs-worktree>/.superpowers/sdd/slot-delta-<round>-<slot>.diff`. Each dispatch sets that slot's sha
to the HEAD it was dispatched against, and a slot not dispatched in a round keeps the sha it had. A
slot for which no last-reviewed sha is held reads the whole `final-review.diff`. Bugbot and Security
read no diff file and are unaffected. Every slot's dispatch prompt names the path it was given and,
for a delta, the sha that delta starts from.

**A required slot whose delta is empty is not dispatched**, and the record states `not re-run —
nothing new since its last read`. Slot 0 reads the whole diff, has no delta, and is never scoped out
by this. **An on-demand slot the operator has not (yet) named at all for this run is never
dispatched under Full mode either** — Full escalates breadth among slots already in play, it never
adds a slot the operator has not asked for; that addition happens only through the explicit-request
check this file's **The two on-demand slots** section states, at the start of any round including a
Full one.

**Escalate automatically** — do not ask, and say why in the record — when the fix touched a file
outside the set named in the findings; the fix altered a capability spec, a migration, or a guard's
behaviour; a targeted re-run surfaced a **new** Critical finding; three or more fix rounds have
already run; or the fix diff exceeds ~150 changed lines **and** adds a new file.

Targeting is a cost optimization, never a coverage waiver: a targeted re-run is never fewer than two
agents, and handoff still requires **zero open findings at any severity** from every agent that has
run.

Union all **open** findings, dedupe by **defect identity — file:line + theme.** *File:line* is the
finding's own recorded location, taken verbatim from the findings table. *Theme* is the finding's
one-sentence Note column, reduced to its own defect noun phrase — the shortest phrase naming what is
wrong, severity words and slot names stripped out.

**Before dispatching the fix subagent**, run

```bash
check-panel-reproducers.sh <worktree> <change>
```

Exit 0 proceeds. Exit 1 covers two classes: a **missing or malformed field** is added before
dispatch; a **rejected reproducer shape** (a shell metacharacter, an absolute path, a `..` segment,
a leading `-`, a URL, a NUL byte) is a **refusal** — the line is recorded **unverifiable** and put to
the operator, never silently rewritten. Exit 2 stops the run.

**For each open finding whose record carries a runnable `finding-reproducer:` command**, run

```bash
run-reproducer.sh <worktree> "<the finding's finding-reproducer: text>"
```

Read its exit code: **0** dispatches the finding; **1** bounces it once, back to the raising slot,
carrying the reproducer's passing output; **2** is a refusal — recorded **unverifiable**, put to the
operator; **3** is a timeout or a detached survivor — recorded **unverifiable**, put to the
operator, with a surviving pid named when the script names one; **4** stops this finding's dispatch
decision entirely — a finding recorded `none — <reason>` is dispatched without a run.

Where a slot **dispatched by `subagent_type`** supplies nothing for a finding, record `none — not
supplied by <slot>` and dispatch the finding unverified. A **general-purpose** slot that supplies
nothing has not supplied a legal exemption: record that omission as its own open finding.

**Once the fix subagent reports, re-run every dispatched finding's reproducer** under the same
constraints and require it now to exit **0**. **The flip alone does not close a finding — the fix's
diff must also touch at least one path the finding named, with a non-comment, non-whitespace
change.** A fix that does not is not a fix: the finding stays open and goes to the operator through
the handback below.

### The fix round mutation-proves what it changed

**This binds the review panel's fix round and not the per-task review's fix** in
`skills/flow/implement.md`.

**Every executable behaviour the fix changed is mutation-proved, not only the test cases the round
adds.** The fix subagent names the executable behaviours its fix changed. **You** then mutate each
one — revert it in a scratch tree, or flip the single value it turns on — confirm an existing test
fails, and restore. An unclear case goes to the operator through the same handback used below.

**Each mutation alters one mechanism.** Where a single revert would also change state a second check
reads, split it into surgical mutations, one per mechanism.

**A surviving mutation is repaired in this round.** Add the test that catches it before the round
closes.

**Record each one in this pass's log entry:**

```text
fix-mutation: <path> — <what was mutated> — <the test that failed>
fix-mutation: <path> — none — <reason>
fix-mutations-total: <n>
```

**These lines go in the pass log entry and never inside the marker block.**
`check-unfinished-work.sh` requires every `finding-status:` marker to occupy one unbroken run of
consecutive lines.

**No line anywhere in the panel record may carry the literal label `finding-status:`,
`findings-total:`, or `finding-reproducer:` outside its own marker use.** Write around it: paraphrase
the label, or break it with a non-word character.

**The parent checks the reported list against the fix diff before the round can close.** Walk every
hunk of the fix diff with a non-comment, non-whitespace change: each one is either covered by a
reported line, or is not an executable behaviour at all. A hunk that removes or weakens a test or an
assertion states in the record what it used to cover and names what still covers that same behaviour
now — checked by running the named covering test against the **pre-fix** code and confirming it
fails.

This binds the fix round every run — the obligation is the round's, not a slot's, so a run where the
operator did not name Bugbot is exactly where the round's own proof is the only mutation reasoning
that happens at all.

**Rebuild the dispatch context bundle before dispatching the fix subagent**, same as above,
overwriting the same path.

**Carry each surviving finding to the fix subagent as a structured block**, not a bare restatement
of its prose: its `F<n>`, the slot that raised it, its severity, its `file:line`, its theme, the
text of its `finding-reproducer:` line, and any bounce already recorded against its defect identity.
**Inline no source excerpt.**

Give the surviving findings to **one** fix subagent as the combined list. Where a finding is
confirmed as a real defect, the fix subagent invokes **superpowers:systematic-debugging** before
writing its fix. **Dispatch it on `DEFAULT_MODEL`** — design.md's `model-default-sonnet` collapses
the panel-fix role's own default onto the single settings-store default, deliberately dropping the
old Opus-panel-fix default `skills/myflow-contracts/model-policy.md` still describes for the retired
per-change fields; that table is stale for `/flow`, per `skills/flow/SKILL.md`'s own note. Record
every pass in `<abs-worktree>/.superpowers/sdd/final-review-panel.md`: mode, which agents ran, why,
the diff path they read, and — when this pass bounced any finding — each bounced finding's defect
identity together with the reproducer output it carried back.

**The fix subagent's own dispatch is recorded too, with `-role panel-fix`:**

```bash
myflow record dispatch begin -change <name> -role panel-fix -model <m> \
  -key panel-fix-<round> -session-token mf-<literal-token> -started-at <ts>
myflow record dispatch end -change <name> -key panel-fix-<round> \
  -session-token mf-<literal-token> -commit <partner-task-sha> -outcome completed -ended-at <ts> \
  -agent-id <id>
```

`-commit` is the task commit the fixup was folded into. `-agent-id` goes on `end`.

A minor finding blocks the handoff exactly as a critical one does. When fix rounds do not converge,
the run hands back to the operator, one finding at a time:

> **`<location>` — <the finding, in one line>. The fix round did not resolve it.**
> - **Take another round on it** *(default, recommended)*
> - **Withdraw it — I'll give the reason** — the reason is recorded on the finding's marker line
> - **Stop the run and hand it back to me**

Only that answer records `withdrawn`, and only with the reason the operator gives.

```bash
myflow stage end -command '/flow' -stage flow.review-panel -outcome completed <name>
```

Once this stage ends clean — zero open findings at any severity, no stale result — continue into
`skills/flow/verify-and-handoff.md`.
