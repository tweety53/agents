# Review panel

Loaded by `skills/flow/SKILL.md` immediately after `skills/flow/implement.md`'s `flow.sdd-tdd`
stage closes, on every implementation run — creating, resumed, or fix. Dispatches `REVIEWERS` —
the roster `skills/flow/SKILL.md`'s **Model resolution** resolves from the settings store, which
that file is canonical for. This file owns dispatch: mapping each resolved id to its slot and
spawning it. Per design.md's `roster-from-settings`, which supersedes `review-panel-fixed-3`
(design.md's `supersedes-review-panel-fixed-3`): there is no fixed roster table and no
diff-size/touched-area trigger table.

```bash
flow stage begin -command '/flow' -stage flow.review-panel -harness <harness> -session-token mf-<literal-token> <name>
```

**Run the citation pre-check before rebuilding the dispatch context bundle below**:

```bash
check-panel-citation-trigger.sh <worktree> <merge-base>
```

**Load `skills/flow-contracts/project-configuration.md`** — the key below is canonical there.

On exit 0, read `<project>/.flow/project.md`'s `## review panel citation check` key (**Guard
resolution**, `skills/flow-contracts/pipeline.md`; the key itself is canonical in **Project
configuration**, `skills/flow-contracts/project-configuration.md`). If declared, run its command
from the apply worktree and capture combined stdout+stderr verbatim to
`<abs-worktree>/.superpowers/sdd/citation-check.md`. On exit 1, or the key absent, skip silently — no
file written, no CITATION CHECK paragraph added below. Exit 2 — the guard could not answer (usage
error, `<worktree>` not a directory or not a git repo, or the merge-base not resolving) — report its
stderr and skip this worktree the same way exit 1 does.

**Never blocks.** The configured command's exit code is read by nobody — a non-zero exit (findings
present) still just writes the file. The real gate stays `flow.verify`'s existing `## lint` run,
unchanged by this step.

**Rebuild the dispatch context bundle at the start of this stage too** — never reused from
`skills/flow/implement.md`'s run. Overwrite the same path:

```bash
mkdir -p <worktree>/.superpowers/sdd
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  > <worktree>/.superpowers/sdd/dispatch-context.md
```

## The roster

Every resolved id maps to one slot, dispatched this run because `REVIEWERS` carries it:

| id | Slot | How to spawn | Model |
|---|------|---------------|-------|
| `primary` | **Primary** — plan alignment + code quality | **superpowers:requesting-code-review** with `final-review.diff` + the plan/spec constraints | `DEFAULT_MODEL` |
| `principles` | **Principles** | general-purpose + `principles-reviewer-prompt.md`; all three principle groups always apply, all three principle groups are always covered <!-- refs-guard:allow --> | `DEFAULT_MODEL` |
| `code-review-low` | **Code review (low)** | general-purpose reviewer briefed for high-confidence defects only, against `final-review.diff` | `DEFAULT_MODEL` |
| `bugbot` | **Bugbot** — defect hunt | `subagent_type: bugbot`, `Diff: uncommitted changes`, `Full Repository Path: <worktree>`, plus the mutation-testing brief below (own throwaway worktree — see **Bugbot's throwaway worktree** below) | none — records `unknown (agent-defined)` |
| `security` | **Security** | `subagent_type: security-review`, same shape as Bugbot | none — records `unknown (agent-defined)` |

`ValidReviewers` in `<agents repo>/stats/internal/store/settings.go` is the id vocabulary this table exhausts —
five entries, never a sixth. `DEFAULT_MODEL` is `skills/flow/SKILL.md`'s **Model resolution** value
for this run. There is no parent-model inheritance and no economy tier. Bugbot and Security are dispatched by
`subagent_type` and carry their own agent definitions — pass them **no** model override, unless the
harness running this stage does not offer that agent type — see below.

### An unspawnable id is substituted, not skipped

Per design.md's `unspawnable-id-substitutes`: a resolved id whose own agent type this harness does
not offer — Bugbot or Security, dispatched elsewhere by `subagent_type` — is dispatched instead as a
**general-purpose** subagent carrying that slot's brief. The panel is never silently reduced by the
harness it happens to run in.

**Check before dispatch, not after a failure.** On Claude Code, the Agent tool's own available agent
types are enumerated in a system-reminder in the conversation before any dispatch; an id's agent type
absent from that listing is unavailable in this harness. Where a harness exposes no such pre-dispatch
listing, attempt the dispatch by `subagent_type` once and treat an immediate rejection naming the
type as unknown or unsupported as the same signal — the same "stated against the mechanism" shape as
**Progress visibility** (`skills/flow-contracts/pipeline.md`).

**The substitute MUST perform mutation testing**, not merely read for defects: for each finding it
raises, it changes the code to prove the finding is real, confirms the change surfaces it, then
reverts. This is the operator's own requirement and is what keeps a substituted slot worth
dispatching — a general-purpose reviewer that only reads is not obviously equivalent to the real
slot; forcing it to prove each finding by mutation is.

**The substitution is recorded, never hidden.** `-slot` still names the slot it stood in for
(`Bugbot` or `Security`); `-model` records the model actually given, never `unknown
(agent-defined)` — that value is correct only for a slot spawned by its own `subagent_type` with its
own agent definition, which a substitute is not. The panel record and the handoff additionally say
in prose which slots were substituted and ran as general-purpose. **A dispatch recorded as Bugbot
that was not Bugbot corrupts the one record that says what reviewed this branch** — the recording
rule above exists to keep that record honest, not merely tidy. A substituted Bugbot's dispatch runs
against the same throwaway worktree treatment as the real slot (**Bugbot's throwaway worktree**),
since it carries the same mutation-testing brief.

**No slot is ever added by diff size, touched area, or any other automatic trigger: the roster is
exactly the resolved list above, and anything beyond it reaches the panel only through an explicit
per-run operator instruction, for that run only.** Check for one at two points: at the start of this stage (has the operator, in this run's own argument or
in the session before this stage, named an id the resolved list does not carry?), and again at the
start of every fix round below — an operator may ask mid-run, after seeing pass 1's result, and that
request adds the slot starting from the round it was made, never retroactively to a pass already
closed. It is never written back to the settings store. Record which slots were added this way and
why (the operator's own words), and record explicitly when none were: "no addition this round — the
resolved list ran alone." — a documentation-, prompt-, or test-only diff with no operator addition
runs the resolved list alone, and that is a correct outcome, said so explicitly, never a silently
skipped review.

**Before writing `final-review.diff`**, run

```bash
generate-relocation-comparison.sh <worktree> <changeRoot> <merge-base>
```

A non-zero exit (`2` — cannot answer) prints one line and the run continues without the comparison
file. This call is never a gate — generation never blocks the panel.

```bash
check-panel-diff-size.sh <worktree> <merge-base>
```

Exit 0 proceeds. Exit 1 puts the choice to the operator, shape per Operator prompts
(`skills/flow-contracts/operator-prompts.md`):

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
flow record dispatch begin -change <name> -role reviewer -slot <slot> -model <m> \
  -agent-id <id> -diff-base <sha> -key panel-<round>-<slot> \
  -session-token mf-<literal-token> -started-at <ts>
flow record dispatch end -change <name> -key panel-<round>-<slot> \
  -session-token mf-<literal-token> -outcome completed -ended-at <ts> -agent-id <id>
```

`-slot` names the slot from **The roster** table above. `-role` is
`reviewer` for every one; `-task` is omitted. `-diff-base <sha>` is the sha a delta-slot's delta
starts from, passed on a slot dispatched against a delta and on no other. `-model` is `DEFAULT_MODEL`
(or this run's override) for every slot except Bugbot and Security dispatched by their own
`subagent_type`, which record `unknown (agent-defined)` — narrowed by **An unspawnable id is
substituted, not skipped**, above: a *substituted* Bugbot or Security slot records the model
actually given, like every other slot.

**On Claude Code, `-agent-id` is the identifier an asynchronous agent launch returns in the parent's
own tool result, at launch.** Never invent one.

**Record a slot's dispatch before recording that slot's findings**, and carry the seq the command
printed — `recorded: dispatch <seq>` — into each of that slot's `flow record finding` calls as
`-dispatch-seq <seq>`.

**Every slot must supply, per finding, a reproducer**: a runnable command that demonstrates the
defect, or the literal exemption form `none — <reason>`. Carry this requirement on every slot's
dispatch prompt.

**Every slot's dispatch prompt also carries the CONTEXT BUNDLE paragraph** — the same one
`skills/flow/implement.md`'s implementer dispatch carries. **When
`<abs-worktree>/.superpowers/sdd/relocation-comparison.md` exists, every slot's dispatch prompt also
names its absolute path**, framed as a review input to audit against the diff — never a substitute
for reading `final-review.diff` itself.

**Every slot's dispatch prompt also carries the FOREGROUND BUILDS paragraph**:

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

**Every slot's dispatch prompt also carries the REPRODUCE, DON'T READ paragraph**:

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST exercise the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.

**Every slot's dispatch prompt also carries the CITATION CHECK paragraph, present only when the
citation pre-check above wrote `<abs-worktree>/.superpowers/sdd/citation-check.md`**, naming its
absolute path alongside `final-review.diff`:

> **CITATION CHECK:** `<abs-worktree>/.superpowers/sdd/citation-check.md` carries this project's
> pre-panel citation scan, captured before your dispatch. It is informational — its own exit code
> was not gating — but a stale citation it reports is worth raising as your own finding if it sits
> in this diff's blast radius.

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
(`skills/flow-contracts/operator-prompts.md`):

> **Slot `<slot>` breached the wall-clock ceiling a second time. How should this proceed?**
> - **Re-dispatch it again**
> - **Proceed without the slot**
> - **Stop the run** *(default, recommended)* — ends at `IN_PROGRESS` with the implementation
>   committed on the branch

A timed-out slot raises no finding, consumes no fix round, and is not a clean result for the final
pass.

### Code review (low)

Code review (low), when dispatched, invokes no skill — unlike Primary and Principles above. Its
findings are ordinary `F<n>` rows, exactly like every other slot's.

### Bugbot's mutation-testing brief

Wherever the panel dispatches Bugbot, its dispatch prompt carries a mutation-testing brief: for
each behaviour the diff changes, mutate it — flip a condition, drop a guard, move a boundary, remove
a branch — and establish whether an existing test fails. A mutation no test catches is a
**surviving mutant**, an ordinary finding that blocks the handoff exactly as any other, unless the
operator withdraws it with a reason.

### Bugbot's throwaway worktree

Bugbot mutates code in place to run its brief; every other slot only reads the diff. Dispatching
Bugbot into the same worktree a reading slot concurrently reads is the KAN-366 collision — a
mutation applied for Bugbot's test is visible to whatever a concurrently dispatched reading slot
reads from `<worktree>` at that moment. Bugbot's dispatch — pass 1, and every fix-round re-run,
a substituted general-purpose Bugbot included (**An unspawnable id is substituted, not skipped**) —
therefore runs against a throwaway worktree, never the shared `<worktree>` the other slots read:

```bash
git -C <worktree> worktree add --detach <worktree>-bugbot-<round> HEAD
git -C <worktree> diff HEAD --binary | git -C <worktree>-bugbot-<round> apply --allow-empty
git -C <worktree> status --porcelain -z | \
  while IFS= read -r -d '' entry; do
    st="${entry:0:2}"; f="${entry:3}"
    [ "$st" = "??" ] || continue
    mkdir -p "<worktree>-bugbot-<round>/$(dirname "$f")"
    cp -a "<worktree>/$f" "<worktree>-bugbot-<round>/$f"
  done
```

`git diff HEAD --binary` — against `HEAD`, not a bare `git diff --binary` — is the same "staged and
unstaged together" semantics `final-review.diff` above already uses, and covers the transplant in
one diff rather than the working-tree-only diff a bare `git diff` produces: a bare `git diff` misses
anything staged, and (independently) fails to reconstruct a rename whose move is already reflected
in the index. `--allow-empty` on the `apply` side makes the sequence a no-op, not a failure, when
there is nothing to transplant — the common case, since task and fix-round work is committed and
`worktree add --detach ... HEAD` already carries every committed change on its own. The
untracked-file loop reads `git status --porcelain -z`, NUL-delimited, into `read -r -d ''` — the
plain-text `awk` form cannot survive git's quote-escaping of a filename with a space or another
special character, and silently drops that file from the copy; the `-z`/NUL form carries the literal
byte string through untouched, regardless of what the filename contains.

Dispatch Bugbot with `Full Repository Path: <worktree>-bugbot-<round>` in place of `<worktree>`.
Remove the copy unconditionally once that dispatch closes — completed, timed out (including after
the wall-clock re-dispatch), or the run stopped:

```bash
git -C <worktree> worktree remove --force <worktree>-bugbot-<round>
```

Findings and reproducers are unaffected: a finding's `file:line` is repo-relative, and every
reproducer still runs against the real `<worktree>` at verification time, never against Bugbot's
copy, exactly as today. Security is **not** isolated this way — nothing in this file requires it to
mutate anything, so it keeps sharing `<worktree>` with the reading slots.

Principles, when dispatched, is the panel's judgment check on *how* the code is built. It reads
`engineering-principles.md` — never a pasted copy — and owns the project's **hard invariants** from
its standards files.

**Resolve `[PRINCIPLES_PATH]` before dispatching the principles slot.** It is the **absolute** path
of `engineering-principles.md` **beside this file** — `skills/flow/`, always. Confirm the file
exists before spawning; if it does not, stop and report rather than dispatching a blind reviewer.

**Resolve `[STANDARDS_PATHS]` before dispatching the principles slot**, from the `## standards` entries in
`<project>/.flow/project.md`, per **Project configuration**
(`skills/flow-contracts/project-configuration.md`). Pass an **empty** value when none resolve.
Record which standards files were passed, or that none resolved.

## Recording findings, and the record's format

**Beside the existing dispatch recording**: as each slot's report comes back and its `flow record
dispatch end` is recorded, write that report byte for byte to
`<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md`. `<round>` is the same value that
round's findings carry on `-round` (`0` initial, `1..n` fix rounds); `<id>` is the resolved reviewer
id, never the slot display name. Every dispatched slot writes one, including a slot that raised
nothing and a slot substituted per **An unspawnable id is substituted, not skipped**. A report that
cannot be captured verbatim still writes its file, carrying the single line `no verbatim report
captured — <reason>`.

**Every finding is a row in the store. The panel record is rendered from those rows.** As each slot
raises a finding, record it — one call, as it is raised:

```bash
flow record finding -change <name> -ref F<n> -round <r> -slot <slot> -severity <sev> \
  -location <file:line> -status open -reproducer <command | none — reason> \
  -dispatch-seq <seq> -note <the finding>
```

`-round` is `0` for the initial panel and `1..n` for a fix round. `-ref` is unique within the change;
the store's own constraint enforces it. **`-note` carries the reviewer's own sentence naming the
defect, not a dispatcher restatement.** Where the reviewer's wording runs long, quote the sentence
that names the defect and leave the rest to the report file. **A fix round updates the finding it
resolved rather than appending a second row:**

```bash
flow record status -change <name> -ref F<n> -status fixed
```

**`-status` carries the whole status text the marker line shows** — a withdrawal passes its reason
with it: `-status 'withdrawn <the operator's reason>'`.

**Render the record when the panel closes** — every slot's result clean, no finding open:

```bash
flow record render -change <name> -kind panel -repo <abs-worktree>
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

`check-unfinished-work.sh` parses no rendered document at all — it queries the store directly
through `flow record findings -change <name> -C <worktree>`, decoded by `jq`, and counts findings
whose status is neither `fixed` nor `withdrawn <reason>`. It reads no table and no marker block;
its own header records why that parser was removed. `validateFindingStatus` in
`<agents repo>/stats/cmd/flow/record.go` already guarantees every stored status is exactly `open`, `fixed` or
`withdrawn <reason>`, so the malformed-shape checks a hand-rolled document parser once needed have
nothing left to catch.

**The table carries no status column, on purpose.** To read a finding's state, look up its `F<n>`
in the marker block.

**A `withdrawn` marker's reason is checked for being there at all.** A finding is recorded `fixed`
by the parent at the fix round's verification step below, never by the fix subagent.

## Panel re-runs

**Pass 1 always runs the resolved roster plus every slot the operator named at this stage's start
that the resolved roster did not already carry.** Only re-runs after a fix are scoped. Record
`FIX_BASE=<task-sha>` — the task's
commit as it stood before this fix round — then, once the fix is folded into that commit via `git
commit --fixup=<task-sha>` and `git rebase --autosquash`, write
`<abs-worktree>/.superpowers/sdd/fix-round-N.diff` from `git diff "$FIX_BASE"..<task-sha>`.

When a review finding requires a code change to a task that is already committed, commit the fix as
`git commit --fixup=<task-sha>`, where `<task-sha>` is the **original** task commit. Immediately
`git rebase --autosquash` to fold it in.

**A clean `git rebase --autosquash` is not evidence the fix survived it.** Where the fixup and the
commit it folds into touch nearby lines, git's 3-way auto-merge can resolve in favour of the
pre-fix side — it exits 0, prints no conflict marker, and leaves no `fixup!` commit behind. The
reproducer rerun and diff check below (**Once the fix subagent reports…**) are what catch this;
they must run against the post-rebase file content, never be satisfied by the fixup commit's
presence or the rebase's own exit code.

| Mode | Who re-runs | Diff they get |
|------|-------------|----------------|
| **Targeted** (default) | Primary, when the roster carries it, as an integration check, plus every agent that raised a finding | `fix-round-N.diff` |
| **Full** (escalation) | Every slot in the resolved roster, plus every added slot already dispatched in an earlier pass of this run | Primary (when it runs) the rewritten `final-review.diff`; every other diff-reading slot its own delta, below |

**A delta is `git diff <the HEAD sha that slot last reviewed> HEAD`**, written to
`<abs-worktree>/.superpowers/sdd/slot-delta-<round>-<slot>.diff`. Each dispatch sets that slot's sha
to the HEAD it was dispatched against, and a slot not dispatched in a round keeps the sha it had. A
slot for which no last-reviewed sha is held reads the whole `final-review.diff`. Bugbot and Security
read no diff file and are unaffected. Every slot's dispatch prompt names the path it was given and,
for a delta, the sha that delta starts from.

**A resolved-roster slot whose delta is empty is not dispatched**, and the record states `not re-run
— nothing new since its last read`. Primary, when it runs, reads the whole diff, has no delta, and
is never scoped out by this. **A slot outside the resolved roster that the operator has not (yet)
named at all for this run is never dispatched under Full mode either** — Full escalates breadth
among slots already in play, it never adds a slot the operator has not asked for; that addition
happens only through the explicit-request check this file's **The roster** section states, at the
start of any round including a Full one.

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

A finding meeting both conditions is recorded closed there and then:

```bash
flow record status -change <name> -ref F<n> -status fixed
```

**The parent records it, never the fix subagent** — the parent is what ran the reproducer and
walked the diff. **Record it per finding, as its verdict is reached, not batched at the round's
end** — an aborted round still leaves every already-verified finding closed. **A finding failing
either condition is left untouched** on `open`, for the handback below.

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

**No line anywhere in the panel record may carry the literal label `finding-status:`,
`findings-total:`, or `finding-reproducer:` outside its own marker use.** Write around it: paraphrase
the label, or break it with a non-word character.

**The parent checks the reported list against the fix diff before the round can close.** Walk every
hunk of the fix diff with a non-comment, non-whitespace change: each one is either covered by a
reported line, or is not an executable behaviour at all. A hunk that removes or weakens a test or an
assertion states in the record what it used to cover and names what still covers that same behaviour
now — checked by running the named covering test against the **pre-fix** code and confirming it
fails.

This binds the fix round every run — the obligation is the round's, not a slot's, so a run where
Bugbot is neither in the resolved roster nor added this run is exactly where the round's own proof
is the only mutation reasoning that happens at all.

**Rebuild the dispatch context bundle before dispatching the fix subagent**, same as above,
overwriting the same path.

**Carry each surviving finding to the fix subagent as a structured block**, not a bare restatement
of its prose: its `F<n>`, the slot that raised it, its severity, its `file:line`, its theme, the
text of its `finding-reproducer:` line, its slot's report path, and any bounce already recorded
against its defect identity. **Inline no source excerpt.**

**Every fix subagent's dispatch prompt also carries the VERBATIM REPORT — THE FACT paragraph**:

> **VERBATIM REPORT — THE FACT:** each finding below names the file its slot's report was written
> to. That file is the reviewer's own report, unedited — read it before you act on the finding.
> The structured block is the dispatcher's summary of it: direction on what to work on, and
> never a source of fact. Where the two disagree the report wins. Where the block asserts
> something the report does not, treat it as unchecked and establish it yourself before building
> on it.

**Every fix subagent's dispatch prompt also carries the FOREGROUND BUILDS paragraph**:

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

Give the surviving findings to **one** fix subagent as the combined list. Where a finding is
confirmed as a real defect, the fix subagent invokes **superpowers:systematic-debugging** before
writing its fix. **Dispatch it on `DEFAULT_MODEL`** — design.md's `model-default-sonnet` collapses
the panel-fix role's own default onto the single settings-store default, deliberately dropping the
old Opus-panel-fix default `skills/flow-contracts/model-policy.md` still describes for the retired
per-change fields; that table is stale for `/flow`, per `skills/flow/SKILL.md`'s own note. Record
every pass in `<abs-worktree>/.superpowers/sdd/final-review-panel.md`: mode, which agents ran, why,
the diff path they read, and — when this pass bounced any finding — each bounced finding's defect
identity together with the reproducer output it carried back.

**The fix subagent's own dispatch is recorded too, with `-role panel-fix`:**

```bash
flow record dispatch begin -change <name> -role panel-fix -model <m> \
  -key panel-fix-<round> -session-token mf-<literal-token> -started-at <ts>
flow record dispatch end -change <name> -key panel-fix-<round> \
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

**Before closing the stage**, run

```bash
check-panel-findings-closed.sh <worktree> <change>
```

Exit 0 proceeds to the stage close below. Exit 1 means a finding still reads `open` in the store —
return to the handback loop above for it. Exit 2 stops the run.

```bash
flow stage end -command '/flow' -stage flow.review-panel -outcome completed <name>
```

Once this stage ends clean — zero open findings at any severity, no stale result — continue into
`skills/flow/verify-and-handoff.md`.
