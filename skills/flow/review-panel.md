# Review panel

Loaded by `skills/flow/SKILL.md` immediately after `skills/flow/implement.md`'s `flow.sdd-tdd`
stage closes, on every implementation run — creating, resumed, or fix. Dispatches the resolved
roster: `REVIEWERS` — `skills/flow/SKILL.md`'s **Model resolution** resolves from the settings
store, which that file is canonical for — when `REVIEW_PANEL_TOGGLE` is `default`, or this run's
decision's `panel.roster` (`<abs-worktree>/.superpowers/sdd/decision.json`, per design.md's **The
`## Decision` block**) when it is `dynamic`. This file owns dispatch: mapping each resolved id to
its slot and spawning it. Per design.md's `roster-from-settings`, which supersedes
`review-panel-fixed-3` (design.md's `supersedes-review-panel-fixed-3`): there is no fixed roster
table and no diff-size/touched-area trigger table.

```bash
flow stage begin -command '/flow' -stage flow.review-panel -harness <harness> -session-token mf-<literal-token> <name>
```

**The pass log is store rows, rendered — never a hand-written file.** Every fact this file records
about a panel run the parent records as it arises with `flow record pass` or `flow record mutation`
(`-change <name> -round <n>`, the round `0` for the initial panel and `1..n` for a fix round).
`flow record render -kind panel` renders them into the panel record's pass-log section under
`<abs-worktree>/.superpowers/sdd/reviews/`, beside the findings.

## Check base movement first

Once per worktree in this run's resolved set (**Resolving a change's worktrees**,
`skills/flow-contracts/worktree-resolution.md`), on every panel run — creating, resumed, or fix:

```bash
BASE="$(resolve-base-branch.sh <worktree>)"
check-base-moved.sh <worktree> "origin/$BASE" <working-notes-merge-base>
```

`<working-notes-merge-base>` is the merge base `skills/flow/implement.md`'s isolate-workspace step
recorded in this run's working notes — never the state file's `worktrees` map, which a creating run
has not written yet. `check-base-moved.sh` performs no fetch of its own; `resolve-base-branch.sh` is
what fetches, so this order — resolve, then check — is load-bearing.

Report every worktree's verdict: `MOVED` with no overlap continues with no prompt; `REFUSE`, an exit 2, or an empty resolved
set stops and asks; an overlap from any worktree asks once for the whole change, shape per Operator
prompts (`skills/flow-contracts/operator-prompts.md`):

> **The base branch has moved and touches paths this change also touched — how should the
> panel proceed?**
> - **Stop — I'll rebase or reorder first** *(recommended)*
> - **Rebase onto `<base>` now, then continue**
> - **Continue — review as is**

**Stop** closes `flow.review-panel` with `-outcome stopped` and leaves the change at its current
state with nothing committed by this stage. **Continue** carries the reported movement into the
handoff and proceeds to the citation pre-check below.

**Rebase** runs `git -C <worktree> rebase origin/$BASE` only in a worktree whose own verdict was
`MOVED` — never one whose verdict was `CLEAR`, even though the prompt above is asked once for the
whole change.

- **Clean** (exit 0): that worktree's working-notes merge base becomes `origin/$BASE`'s resolved
  tip at rebase time; every later `<merge-base>` this file and the `worktrees` map
  `skills/flow/verify-and-handoff.md` writes read the working notes, so nothing else needs
  plumbing. Re-run `check-base-moved.sh` once more against the new value; a fresh overlap re-offers
  the prompt above rather than looping silently. The rebase clears every slot's held last-reviewed
  sha, so a re-run after it reads the whole `final-review.diff` under **Panel re-runs**' existing
  no-held-sha rule below. No re-verification runs here — the panel reads the rebased tree, and
  `flow.verify` follows. End the clean-rebase report with the fixed literal:

  > If this change's verification compares against a recorded baseline, recapture it now — a
  > proof taken against the pre-rebase base is void.

- **Conflict** (non-zero exit): never auto-abort, and never resolve the conflict — by editing the
  conflicting files or otherwise. Leave the worktree mid-rebase exactly as `git rebase` left it,
  report the conflicting file(s) from `git status`, and hand off `git -C <worktree> rebase
  --continue` (after the **operator** resolves it) or `git -C <worktree> rebase --abort` as the
  next manual step. State stays as it was; this stage stops here and closes the mark `stopped`.

**Everything from the citation pre-check to the throwaway worktrees and the `[PRINCIPLES_PATH]`
and `[STANDARDS_PATHS]` resolution below is one Bash call**, under **Turn discipline**
(`skills/flow/implement.md`): every verdict printed, and only the dispatch depends on them — an
over-cap exit, a `REFUSE`, an absent principles file or an operator prompt is read off that call's
output and handled before any slot launches. The base-movement check above is the one exception:
its rebase branch changes what every later step reads, so it runs and is read first.

**Run the citation pre-check before rebuilding the dispatch context bundle below**:

```bash
check-panel-citation-trigger.sh <worktree> <merge-base>
```

On exit 0, run `project-get.sh <worktree> "review panel citation check"` (**Guard resolution**,
`skills/flow-contracts/pipeline.md`; the key itself is canonical in **Project configuration**,
`skills/flow-contracts/project-configuration.md`). If declared, run its command
from the apply worktree and capture combined stdout+stderr verbatim to
`<abs-worktree>/.superpowers/sdd/citation-check.md`. `project-get.sh` exit 1 is the "key absent —
skip silently" case just named; exit 2 is reported and the worktree skipped, the same way the
guard's own exit 2 below is. On exit 1, skip silently — no file written, no CITATION CHECK
paragraph added below. Exit 2 — the guard could not answer (usage error, `<worktree>` not a
directory or not a git repo, or the merge-base not resolving) — report its stderr and skip this
worktree the same way exit 1 does.

**Never blocks.** The configured command's exit code is read by nobody — a non-zero exit (findings
present) still just writes the file. The real gate stays `flow.verify`'s existing `## lint` run,
unchanged by this step.

**Rebuild the dispatch context bundle at the start of this stage too** — never reused from
`skills/flow/implement.md`'s run. Overwrite the same path:

```bash
mkdir -p <worktree>/.superpowers/sdd
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  <worktree>/.superpowers/sdd/dispatch-context.md "" <canonical-worktree> <shape>
```

Report the script's stderr line (`bundle unchanged — reusing …` or `bundle rebuilt — …`) as part
of this stage's own reporting.

## The roster

Every resolved id maps to one slot, dispatched this run because the resolved roster (`REVIEWERS`,
or the decision's `panel.roster` on `dynamic` — see the opening paragraph above) carries it — the
spawn column names the `REVIEW_PANEL_TOGGLE: default` spawn, which the `dynamic` paragraph below
overrides:

| id | Slot | How to spawn |
|---|------|---------------|
| `primary` | **Primary** — plan alignment and senior code review | `flow-review` + `primary-reviewer-prompt.md`: `final-review.diff` against `proposal.md`, `design.md` and each task's `**Files:**`/`**Tests:**`/`**Commit:**` fields in `tasks.md`, plus code quality, architecture, testing and production readiness |
| `principles` | **Principles** | `flow-review` + `principles-reviewer-prompt.md`; all three principle groups always apply <!-- refs-guard:allow --> |
| `code-review-low` | **Code review (low)** | `flow-review` reviewer briefed for high-confidence defects only, against `final-review.diff` |
| `simple-reviewer` | **Simple reviewer** — small class's compact-roster code-quality slot | `flow-review` reviewer briefed for high-confidence defects only, against `final-review.diff`, by `skills/flow/simple-reviewer-prompt.md`, own throwaway worktree copy per repository (see **The throwaway worktree** below) |
| `bugbot` | **Bugbot** — defect hunt | `flow-review` + `bugbot-reviewer-prompt.md`, own throwaway worktree copy per repository (see **The throwaway worktree** below) |
| `security` | **Security** | `flow-review` + `security-reviewer-prompt.md` |
| `mutation` | **Mutation** — sabotage-proofing | `flow-review` + the mutation-testing brief below, own throwaway worktree copy per repository (see **The throwaway worktree** below) |

**A subagent-facing file is passed by absolute path, never read into this context.** Superpowers'
`primary-reviewer-prompt.md` (Primary), `principles-reviewer-prompt.md` and
`engineering-principles.md` (Principles),
`bugbot-reviewer-prompt.md` (Bugbot), `security-reviewer-prompt.md` (Security),
`simple-reviewer-prompt.md` (Simple reviewer), and
`<project>/.flow/project.md`'s standards files are inputs to the slot that reads them; the
dispatcher resolves their paths, confirms each exists, and names them in the prompt.

`ValidReviewers` in `<agents repo>/stats/internal/store/settings.go` is the id vocabulary this table exhausts —
seven entries, never an eighth. `DEFAULT_MODEL` is `skills/flow/SKILL.md`'s **Model resolution** value
for this run. There is no parent-model inheritance and no economy tier. Every slot in this table,
Bugbot and Security included, is dispatched on `flow-review` (`agents/flow-review.md`, on
`REVIEW_PANEL_TOGGLE: default`) and carries the same model rule. `flow-review` is a definition
this repository owns, whose `tools:` allowlist omits `Agent` — **No forking** below is backed by a
capability the slot structurally does not have, not by the NO DELEGATION paragraph alone;
`general-purpose` is a harness-provided type whose tool set cannot be restricted.

**Check for an operator-named id at two points**: at the start of this stage (has the operator, in
this run's own argument or in the session before this stage, named an id the resolved list does not
carry?), and again at the start of every fix round below — an operator may ask mid-run, after seeing
pass 1's result, and that request adds the slot starting from the round it was made, never
retroactively to a pass already closed. It is never written back to the settings store. Record which
slots were added this way and why (the operator's own words) with `flow record pass -round <round>`,
and record explicitly when none were: "no addition this round — the resolved list ran alone."

**On `REVIEW_PANEL_TOGGLE` `dynamic`**, model and effort belong to the dispatch, not the slot:
each entry of the decision's `panel.dispatches` carries its `slots` and its own `model` and
`effort`, and every slot in it runs on that pair — the dispatch's `subagent_type` is
`flow-<effort>` — the effort comes from the definition, the model from the Agent tool's own
`model` parameter, passed explicitly on the dispatch — and both `model` and `-effort` are recorded,
per design.md's `agent-definitions-universal-handshake`. The roster carries no per-slot model. A compact roster
(the decision's `panel.compact`) is recorded with `flow record pass -round 0 -note 'roster: compact — <rolled value>'`; a full
roster records `roster: full`.

### Experimental slot

When the decision's `panel.roster` carries an entry whose `slot` starts `exp-` — at most one, per
design.md's **The rolls** — it is dispatched once, in pass 1 alongside the rest of the roster,
exactly like any other slot in **The roster** table above, carrying the same paragraphs every
slot's dispatch already carries above.

Its prompt is not `principles-reviewer-prompt.md` or any other fixed template: it is the file the
roster entry names in `prompt` — `skills/flow/experimental/<name>.md` inside the agents repo, never
a path inside the project worktree — read by **absolute** path and substituted into the dispatch
prompt the same way `[PRINCIPLES_PATH]` is resolved for Principles. Confirm the file exists before
dispatching; an absent file at dispatch time (the roster was decided against a prompt that has since
moved) is reported and this slot dropped from this run, never dispatched against nothing.

Its `-slot` on the dispatch record, and every `flow record finding -slot` this slot raises, is the
full `exp-<name>` id verbatim — never shortened to `experimental` or to `<name>` alone. Its report
file is `<abs-worktree>/.superpowers/sdd/panel-report-<round>-exp-<name>.md`, the same
`panel-report-<round>-<id>` shape every slot's REPORT FILE paragraph already names with `<id>`
substituted. The rendered panel record's Slot column therefore shows the `exp-` id unchanged, so the
prefix survives into the archive per design.md's `exp-slot-prefix`.

It is a diff-reading slot like Primary, Principles, Code review (low) and Mutation: **Panel
re-runs** below governs it unchanged. **The docs-only reduction** below still narrows a
docs-only branch to `primary` alone: the experimental slot is never part of that reduced roster, and
is dispatched again only if a later round's docs-only guard reclassifies the branch off the
reduction.

It runs at most once per change, whether or not the roster is `compact` — the experimental roll and
the compact roll are independent per design.md's **The rolls** — and never at all when
`REVIEW_PANEL_TOGGLE` is `default`, or when the decision recorded `experimental: none available`.

Per **Bundled dispatch** below, it joins whichever group has room, last among the reading passes;
when neither group has room for a third role it is skipped and recorded with
`flow record pass -round <round> -note 'experimental: skipped — bundle cap'` (design.md's
`exp-skipped-over-cap`) rather than displacing a persistent role.

**Before writing `final-review.diff`**, run

```bash
generate-relocation-comparison.sh <worktree> <changeRoot> <merge-base>
```

A non-zero exit (`2` — cannot answer) prints one line and the run continues without the comparison
file. This call is never a gate — generation never blocks the panel.

```bash
check-panel-diff-size.sh <worktree> <merge-base>
```

once per worktree in the resolved set, unchanged. **The gating count is the sum across
worktrees** — one slot now reads every section — and the over-cap prompt names the sum and each
worktree's own count (design.md's `cap-sum-across-worktrees`).

Exit 0 proceeds. Exit 1 puts the choice to the operator, shape per Operator prompts
(`skills/flow-contracts/operator-prompts.md`):

> **The panel diff measured `<count>`, over the `<cap>` cap. How should this proceed?**
> - **Proceed with the panel anyway** *(default, recommended)*
> - **Stop and split the change** — ends the run at `IN_PROGRESS` with the implementation committed
>   on the branch

Exit 2 stops the run. Record the measured count, the cap in force, and the operator's answer where
one was given with `flow record pass -round <round>` on **every** run, including
exit-0 runs.

Then run

```bash
check-panel-docs-only.sh <worktree> <merge-base>
```

### The docs-only reduction

Per design.md's `docs-only-reduces-to-primary` (narrowing `roster-from-settings`): **exit 0 from
every worktree in the resolved set — every path this branch touched, committed since the merge
base, staged or unstaged, ends `.md` or `.mdc` — reduces pass 1 to `primary` alone**, plus every
slot the operator's per-run instruction
named at this stage's start. Every other resolved slot is recorded with
`flow record pass -round 0 -note 'not dispatched — docs-only reduction: <slot>'`.
`primary` is the reduced roster even when the resolved list does not carry it — the
same shape **Model resolution** (`skills/flow/SKILL.md`) already defines for an empty store list.
This reduction applies to a dynamic roster unchanged: it still narrows to `primary` alone, on the model and
effort of the decided dispatch that carried `primary` — one dispatch, never bundled.

**Exit 1 runs the resolved roster unchanged**; the first non-documentation path any worktree's run
printed is recorded beside the verdict. An empty touched-path set is exit 1 too. One worktree at
exit 1 or 2 runs the resolved roster unchanged for the whole change.
**Exit 2 reports the guard's stderr and runs the resolved roster unchanged**: an unanswered
question never reduces a panel.

Record the verdict, the printed path where there is one, and the roster actually dispatched with
`flow record pass -round <round>` on **every** run, beside the diff-size
fields above.

**This is the one automatic reduction, and it only ever removes.** No slot is ever added by diff
size, touched area, or any other automatic trigger: beyond the reduced or resolved roster,
anything reaches the panel only through an explicit per-run operator instruction, for that run
only.

An inline run checks the context ceiling (**Inline — the parent implements**,
`skills/flow/implement.md`) before this pass 1 dispatch begins.

Write `<abs-worktree>/.superpowers/sdd/final-review.diff` (the canonical worktree's) once per round from **every**
worktree in the change's resolved set (**Resolving a change's worktrees**,
`skills/flow-contracts/worktree-resolution.md`), in resolved order — one first line naming the
file's own semantics, then each worktree's section, opened by a header naming it and its own
working-notes merge base, followed by that worktree's `git diff <merge-base>` (staged and
unstaged):

```sh
: > <abs-worktree>/.superpowers/sdd/final-review.diff
printf '# final-review.diff — working tree vs merge-base, unstaged changes included\n' \
  >> <abs-worktree>/.superpowers/sdd/final-review.diff
# for each <worktree> in the resolved set, in order:
printf '# worktree: %s — merge base %s\n' "<worktree>" "<merge-base>" \
  >> <abs-worktree>/.superpowers/sdd/final-review.diff
git -C <worktree> diff <merge-base> >> <abs-worktree>/.superpowers/sdd/final-review.diff
```

A single-worktree change writes the same shape with one header. The first line is the file's
own answer to the reviewer who reads it as anything narrower: it is a plain working-tree diff
against the merge base, so work still unstaged or uncommitted is already in it — KAN-459's F37
was a false positive from reading it otherwise. Then dispatch the round's
`panel.dispatches` in the canonical worktree, each reading the whole combined file; a role is
never dispatched once per worktree: one pass reads every worktree's section, so a seam between two repositories is in one pass's view
(design.md's `combined-diff-per-round`). **Bundled dispatch** below states how the roster is
grouped into those dispatches.

### Bundled dispatch

**At most two review dispatches per round, each carrying one to three roles**, on both
`REVIEW_PANEL_TOGGLE` values and in both execution modes (design.md's
`two-dispatch-cap-everywhere`). A dispatch carrying one role covers that role alone; a roster the
two dispatches cannot hold shrinks to what they hold.

**Grouping.** On `dynamic`, the decision's `panel.grouping` is `static` — the class's row in
design.md's **Bundled dispatch › Grouping** table, unchanged, no override — or `free` — the
planner's own grouping within the ≤2 × ≤3 cap, recorded as `panel.grouping_reason`. On `default`,
the settings-store roster is grouped deterministically by the same static logic, no roll and no
planner: reading roles (`primary`, `principles`, `security`, `code-review-low`) fill the first
dispatch in that order up to three, the rest and the mutating roles (`bugbot`, `mutation`) the
second, up to three; a list the two cannot hold is truncated in store order, and the truncation is
recorded with `flow record pass -round <round>`.

**One `dispatches` row per bundle** — the same `flow record dispatch begin`/`end` pair below, with
`-slot` the bundle's roles `+`-joined in roster order (`primary+principles+security`) and
`-model`/`-effort` the bundle's own, from the decision's `panel.dispatches` entry. Every finding still records its own single role in `-slot`, with the
bundle's `-dispatch-seq`. A one-role dispatch is unchanged from today.

**The bundle prompt** carries the shared paragraphs — CONTEXT BUNDLE, WORKTREES, TOOLS, FOREGROUND
BUILDS, REPRODUCE DON'T READ, CITATION CHECK, MODEL HANDSHAKE, the reproducer rule — once, then one
**PASS `<id>`** section per role in roster order, each carrying exactly the brief that role's solo
dispatch carries above and its own REPORT FILE line naming `panel-report-<round>-<id>.md`. Mutating
roles (`mutation`, `bugbot`) are always the last passes of a bundle and still work in their
throwaway copies (**The throwaway worktree** below); simple-reviewer works in its own throwaway
copy too, read-only, wherever one is made for it; the reading passes before them read the shared
`<worktree>`. The return message carries one findings summary per role under a heading naming the
role; the parent records each finding under that role.

Every bundle prompt also carries this paragraph verbatim:

> **INDEPENDENT PASSES:** each pass starts from `final-review.diff` and the code, never from an
> earlier pass's report or conclusions. Do not cite, defer to, or skip a defect because an earlier
> pass raised it — if it sits in this pass's angle, raise it again under this pass. Write each
> pass's report file before beginning the next pass.

**No de-duplication across roles**: the same defect raised by two passes is two `F<n>` rows.

**Re-runs are re-grouped by the same grouping**, carrying only the roles re-running this round — a
group whose other members are clean dispatches with its re-running members only.

The rendered panel record's pass-log section and the `IN_PROGRESS` handoff's `Panel:`
line name the dispatches as `+`-joined groups (`primary+principles · code-review-low+mutation`).

**Every slot's dispatch is recorded**, the same pair section 4 of `skills/flow/implement.md`
records for an implementer:

```bash
flow record dispatch begin -change <name> -role reviewer -slot <slot|slot+slot+slot> -model <m> -effort <e> \
  -agent-id <id> -diff-base <sha> -key panel-<round>-<that slot> \
  -session-token mf-<literal-token> -started-at <ts>
flow record dispatch end -change <name> -key panel-<round>-<that slot> \
  -session-token mf-<literal-token> -outcome completed -ended-at <ts> -agent-id <id>
```

Every dispatch of a round launches in one message; every `begin` is recorded in the next Bash
call, one call for all; the round's wait is one call whose condition is `test -s` on every
launched pass's report file; every `end` is recorded in one call once they all exist (**Turn
discipline**, `skills/flow/implement.md`). A dispatch whose report never appears within its
ceiling takes the breach path under **No forking, and a wall-clock ceiling on every slot** below.

`-slot` names the dispatch's roles from **The roster** table above. `-role` is
`reviewer` for every one; `-task` is omitted. `-diff-base <sha>` is passed on a dispatch whose
roles are all reading against a delta and on no other; it takes one
sha, so it carries the **canonical worktree's** held last-reviewed sha, and the
panel record names every worktree's sha beside the delta path (design.md's
`diff-base-canonical-sha`). `-model` is `DEFAULT_MODEL` (or this run's override) on
`REVIEW_PANEL_TOGGLE` `default` and the dispatch's own model from the decision's
`panel.dispatches` on `dynamic` — bundled or one-role alike, no exception. `-effort` likewise:
`default` on `REVIEW_PANEL_TOGGLE` `default`, and the dispatch's own effort on `dynamic`.

**On Claude Code, `-agent-id` is the identifier an asynchronous agent launch returns in the parent's
own tool result, at launch.** Never invent one.

**Record a slot's dispatch before recording that slot's findings**, and carry the seq the command
printed — `recorded: dispatch <seq>` — into each of that slot's `flow record finding` calls as
`-dispatch-seq <seq>`.

**Every slot must supply, per finding, a reproducer**: a runnable command that demonstrates the
defect, or the literal exemption form `none — <reason>`. **The exemption form is available to Minor
findings only: an Important-severity finding must carry a runnable command** — one that
`check-panel-reproducers.sh` accepts and the parent can run — and the guard rejects the exemption
at Important (KAN-503). A demonstrating command needing a pipe, a
quote, a glob or any other shell metacharacter is written as a script rather than abandoned — the
guards refuse a metacharacter in the recorded line, never one inside a script. The slot writes it
to `<abs-worktree>/.superpowers/sdd/reproducers/<round>-<id>-<n>.sh` — `<round>` this round's
number, `<id>` the slot's own resolved reviewer id, `<n>` that slot's own 1-based finding index —
gives it a shebang and `chmod +x`, and records that same path, the path relative to the worktree —
not the `<abs-worktree>/`-prefixed form above; `run-reproducer.sh` refuses an absolute token.
Bugbot, Mutation and Simple reviewer write theirs into the canonical worktree, never their own
`<worktree>-<slot>-<round>` copy, which is removed the moment their dispatch closes. The parent
records the path the slot supplied verbatim — there is no rename step. Carry this requirement on
every slot's dispatch prompt.

**Every slot's dispatch prompt also carries the CONTEXT BUNDLE paragraph** — the same shape
`skills/flow/implement.md`'s implementer dispatch carries; for every worktree in this run's
resolved set, naming that worktree's own five-argument bundle
`<abs-worktree>/.superpowers/sdd/dispatch-context.md`, one path each. **For every worktree whose
`<abs-worktree>/.superpowers/sdd/relocation-comparison.md` exists, every slot's dispatch prompt also
names its absolute path**, framed as a review input to audit against the diff — never a substitute
for reading `final-review.diff` itself.

**Every slot's dispatch prompt also carries the FOREGROUND BUILDS paragraph**:

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

**Every slot's dispatch prompt also carries the TOOLS paragraph**:

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

**Every slot's dispatch prompt also carries the NO DELEGATION paragraph**:

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the parent's closed list (**Dispatch sites — the
> parent's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

**Every slot carries the MODEL HANDSHAKE paragraph** — no exception:

> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.

The dispatcher compares that line against the model this slot was given and applies **The
handshake** (`skills/flow/implement.md`, **The parent orchestrates directly**), unchanged: a first mismatch
is a fallback plus one retry under `panel-<round>-<slot>-retry`; a second is a fallback plus
`## Question`.

**Every slot's dispatch prompt also carries the REPRODUCE, DON'T READ paragraph**:

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST exercise the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.

**Every slot's dispatch prompt also carries the REPORT FILE paragraph**, with the round and the
slot's resolved id substituted:

> **REPORT FILE:** write your full report to
> `<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md` before you end your turn — every
> finding with its `file:line`, severity, the sentence naming the defect, and its reproducer. Your
> return message is the findings summary; the file is the record.

**Every slot's dispatch prompt also carries the WORKTREES paragraph**, listing the resolved set
and, when it holds more than one worktree, the qualification rule:

> **WORKTREES:** this change spans `<abs-worktree-1>`, `<abs-worktree-2>`, …; `final-review.diff`
> is sectioned by worktree, each section headed `# worktree: <path> — merge base <sha>`. When more
> than one is listed, prefix every finding's `file:line` with that worktree's basename
> (`gymie-frontend:src/Foo.tsx:42`) and write its reproducer to run from that worktree.

**Every slot's dispatch prompt also carries the CITATION CHECK paragraph, present for every
worktree whose citation pre-check above wrote `<abs-worktree>/.superpowers/sdd/citation-check.md`,
one path each**, naming its absolute path alongside `final-review.diff`:

> **CITATION CHECK:** `<abs-worktree>/.superpowers/sdd/citation-check.md` carries this project's
> pre-panel citation scan, captured before your dispatch. It is informational — its own exit code
> was not gating — but a stale citation it reports is worth raising as your own finding if it sits
> in this diff's blast radius.

### No forking, and a wall-clock ceiling on every slot

No panel slot is dispatched onto a skill or agent that forks its own background agent — a forked
agent reports to nobody the dispatcher is tracking. Repair it by dispatching the slot on a shape
that reports back directly; never drop the slot.

Every panel slot carries a 15-minute wall-clock ceiling from its dispatch. The dispatcher
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

### The mutation-testing brief

Wherever the panel dispatches Bugbot or Mutation, the dispatch prompt carries a mutation-testing
brief: for each behaviour the diff changes, mutate it — flip a condition, drop a guard, move a
boundary, remove a branch, move an interaction off its target, overlay an earlier commit's tree and
run the tests — and establish whether an existing test fails. A mutation only counts once its edit
landed: every mutation must confirm its edit landed before the tests run — the target changed
where the slot intended, not nowhere and not somewhere else. An edit that never applied must be
redone with a working mechanism: it is a refusal, never a **surviving mutant**, and it never buys
a test. A mutation no test catches is a
**surviving mutant**, an ordinary finding that blocks the handoff exactly as any other, unless the
operator withdraws it with a reason.

### The throwaway worktree

Bugbot and Mutation both mutate code in place to run their brief; simple-reviewer reads without
mutating; every other slot only reads the diff. Dispatching a mutating slot into the same worktree
a reading slot concurrently reads is the KAN-366
collision — a mutation applied for one slot's test is visible to whatever a concurrently dispatched
reading slot reads from `<worktree>` at that moment. **Independent multi-slot detection is the
panel's core signal, preserved deliberately rather than treated as incidental (KAN-503)**: each of
the three defect-hunting slots — Bugbot's and Mutation's dispatch, pass 1 and
every fix-round re-run, both carrying the mutation-testing brief (Bugbot's own copy is
`bugbot-reviewer-prompt.md`'s), and Simple reviewer's read-only pass beside them — therefore runs
against its own throwaway worktree, never the
shared `<worktree>` the other slots read, so every slot's findings are raised against the same
pristine snapshot and no slot's view ever contains another slot's work:

**The parent creates and removes every throwaway copy itself, in its own Bash calls — never a
subagent.** Run the sequence below once per worktree in the resolved set per slot, producing one
`<worktree>-<slot>-<round>` per repository per slot — `<slot>` is the id (`bugbot`, `mutation` or
`simple-reviewer`), so a roster carrying all three produces three copies per repository per round.

```bash
git -C <worktree> worktree add --detach <worktree>-<slot>-<round> HEAD
git -C <worktree> diff HEAD --binary | git -C <worktree>-<slot>-<round> apply --allow-empty
git -C <worktree> status --porcelain -z | \
  while IFS= read -r -d '' entry; do
    st="${entry:0:2}"; f="${entry:3}"
    [ "$st" = "??" ] || continue
    mkdir -p "<worktree>-<slot>-<round>/$(dirname "$f")"
    cp -a "<worktree>/$f" "<worktree>-<slot>-<round>/$f"
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

Dispatch each slot present in this round's roster **once**, its prompt listing every copy made for
that slot as the repository paths to mutate and test in — for simple-reviewer, to read and grep
in — in place of `<worktree>` (design.md's
`bugbot-security-one-dispatch`).
Remove every copy unconditionally once that slot's dispatch closes — completed, timed out
(including after the wall-clock re-dispatch), or the run stopped:

```bash
# for each copy:
git -C <worktree> worktree remove --force <worktree>-<slot>-<round>
```

Findings and reproducers are unaffected: a finding's `file:line` is repo-relative, and every
reproducer still runs against the real `<worktree>` at verification time, never against any slot's
throwaway copy, exactly as today. Security is **not** isolated this way — nothing in this file requires it to
mutate anything, so it keeps sharing `<worktree>` with the reading slots. It too is dispatched
once, its prompt naming every worktree in the resolved set.

Principles, when dispatched, is the panel's judgment check on *how* the code is built. It reads
`engineering-principles.md` — never a pasted copy — and owns the project's **hard invariants** from
its standards files.

**Resolve `[PRINCIPLES_PATH]` before dispatching the principles slot.** It is the **absolute** path
of `engineering-principles.md` **beside this file** — `skills/flow/`, always. Confirm the file
exists before spawning; if it does not, stop and report rather than dispatching a blind reviewer.

**Resolve `[STANDARDS_PATHS]` before dispatching the principles slot**, from the entries
`project-get.sh <worktree> standards` prints (exit 1: none declared), resolved per the entry-form
and containment rule the `[STANDARDS_PATHS]` step of `skills/flow/principles-reviewer-prompt.md`
carries. Pass an **empty** value when none resolve.
Record which standards files were passed, or that none resolved.

## Recording findings, and the record's format

**The slot writes its own report.** Each slot writes
`<abs-worktree>/.superpowers/sdd/panel-report-<round>-<id>.md` itself, per the REPORT FILE paragraph
its prompt carries — `<round>` the same value that round's findings carry on `-round` (`0` initial,
`1..n` fix rounds), `<id>` the resolved reviewer id, never the slot display name. As each slot's
`flow record dispatch end` is recorded, confirm the file exists and is non-empty (`test -s`); when
it is not, write it yourself carrying the single line `no verbatim report captured — <reason>`.
Every dispatched slot ends up with one, a slot that raised nothing included. **Never re-emit a
slot's report from this context** — record its `F<n>` rows and cite the file, per **Read
discipline**'s never-`cat`-a-report rule (`skills/flow/implement.md`).

When two independently dispatched slots raise the same defect — the code-quality ground Primary
and Simple reviewer or Code review (low) now overlap, per `primary-reviewer-prompt.md`'s **Do
Not** — the dispatcher records it once, under a `+`-joined `-slot` value naming both, rather than
as two `F<n>` rows.

**Every finding is a row in the store. The panel record is rendered from those rows.** The parent
records every finding itself, never a subagent. Every
finding a round raised is recorded in one Bash call, one `flow record finding` per finding:

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

**The table carries no status column, on purpose.** To read a finding's state, look up its `F<n>`
in the marker block.

**A `withdrawn` marker's reason is checked for being there at all.** A finding is recorded `fixed`
by the parent at the fix round's verification step below, never by the fix subagent.

## Panel re-runs

**Every round this stage dispatches after pass 1 — each fix-round re-run and the `full` rerun
policy's final pass — opens by running **Check base movement first** again: the same
per-worktree `resolve-base-branch.sh` then `check-base-moved.sh` pair, with that section's
verdicts, operator prompt, rebase and conflict handling unchanged.** **Continue** at a round
boundary proceeds into the round's own remaining steps — the citation pre-check that follows the
entry check is an entry step, and the round does not re-run it. The entry check ran once,
before pass 1; a base that moves while earlier rounds ran would otherwise reach the final round
— and then integrate — unchallenged, its conflict surfacing only after review has closed. A
conflict found here surfaces while the panel is still active and the operator is already
engaged.

**Pass 1 runs the roster **The docs-only reduction** chose — the resolved roster, or `primary`
alone on a docs-only branch — plus every slot the operator named at this stage's start that it
did not already carry.** Only re-runs after a fix are scoped. Record
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

**Which slots re-run, and on what, follows from the severities the round raised — never from a
mode table, a trigger list, or a round count.**

**Every Critical and Important goes to the fix; a Minor is deferred by default and triggers no
re-run either way.** Every Critical and Important the round raised goes to the fix subagent below,
closed by the verification that follows it — the reproducer re-run exits 0 *and* the fix diff
touches a path the finding named. For each Minor, the dispatcher decides before the fix goes out:
`flow record status -change <name> -ref F<n> -status 'deferred <reason>' -category <doc-only|pre-existing|cosmetic|coverage-gap|out-of-scope|other>` is the default — the category naming the mechanism the reason clause states, so the deferred-Minor rate is a query rather than a hand-read; **fix**
it inline only when it is trivially easy — confined to the lines the finding names, needs no new
test, and needs no judgment call. The rough 90%+ target for deferred Minors is guidance, never
computed — no counter, no draw, fixing every Minor is not the goal. A Minor fixed inline under this
bar needs no dedicated re-review or re-verification pass of its own — the triviality that qualified
it for the inline fix is also why it needs none; the existing guard/test run covering the touched
lines is sufficient. When every finding the round raised was Minor, no slot re-runs: proceed to
`check-panel-findings-closed.sh` and the stage close. A fixed finding that fails verification takes
the handback below, and that loop re-runs no slot either.

**A deferral's reason is one clause naming the mechanism — never a rationale essay, in the store
row or in the round's output.**

**When the round raised anything above Minor, re-run on deltas.** A slot's last-reviewed sha is
held **per slot per worktree**: each dispatch sets that slot's sha in every worktree to the HEAD it
was dispatched against, and a slot not dispatched in a round keeps the shas it had. A delta is
`<abs-worktree>/.superpowers/sdd/slot-delta-<round>-<slot>.diff` (the canonical worktree's), combined with the same
per-worktree sections as `final-review.diff`, without its semantics first line — one `# worktree:`
header per worktree, followed by that worktree's `git
diff <held-sha> HEAD`; a worktree in which the slot holds no sha contributes its whole `git diff
<merge-base>` section. Every slot's dispatch prompt names the path it was given and, for a delta,
each worktree's starting sha. **Check base movement first** above clears every slot's held sha in
the rebased worktree on a clean rebase, whether taken at panel entry or at a round boundary, so
that worktree's section falls under the no-held-sha rule in the next round. Then:

- **a slot re-runs only when it raised a finding in the previous round, or the previous round
  raised a new Critical** — every slot in the resolved roster, Primary included, and every
  operator-added slot already dispatched in an earlier pass of this run, on that one rule. A slot
  that raised nothing keeps the result it has; the round's own mutation-proof (below) covers what
  the fix changed;
- **a diff-reading slot that re-runs reads its delta**; Bugbot, Mutation and Security read no diff
  file and re-run in their pass-1 shape, throwaway worktree included. **A diff-reading slot whose
  delta is empty in every worktree is not dispatched** — record `not re-run —
  nothing new since its last read` with `flow record pass -round <round>`;
- **a slot the operator has not named for this run is never added here** — that addition happens
  only through the explicit-request check **The roster** states, at the start of any round.

**From a change's third fix round on, a fix round is scoped.** A re-running diff-reading slot
reads the round's `fix-round-N.diff` plus the sites of every finding an earlier round raised —
each site opened at its recorded `file:line` in the current tree — in place of its held-sha
delta; the re-run rule above is unchanged, and so is everything the round's own mutation-proof
covers. The scoping exists because a round that re-reads a growing fix diff regress-checks by
volume, not by site. A scoped round no longer reads
the branch, so a run that reached one closes with the final whole-branch pass **Rerun policy
`full`** adds — reserved for catching independent issues, which is what caught that change's
round-6 real bugs.

**Rerun policy `full`** — the decision's `panel.rerun` on a `big` class — keeps every rule above for
every fix round and adds one final pass after the last fix round closes clean: every slot in the
roster re-reads the whole `final-review.diff` (Bugbot and Mutation in their pass-1 shape). A finding
from that final pass opens an ordinary fix round under the rules above; the final pass then repeats
once that round closes clean. **Rerun policy `delta`** — `small` and `regular`, and every run on
`REVIEW_PANEL_TOGGLE` `default` — is the section above as it stands: no added final pass, beyond
the one the scoped-round rule above requires of a run that reached a third fix round.

**The cap check on a re-run** is `check-panel-diff-size.sh <worktree> <sha> <cap>` once per
worktree per **distinct** held sha among the diff-reading slots dispatched this round (two slots
sharing a sha in a worktree need one call there, not two); a slot with no held sha in a worktree
counts from that worktree's merge base. **The gating count is the largest per-slot sum across
worktrees** — the largest single combined read any one slot this round faces — and an exit-1
result from a call contributing to it puts the over-cap choice to the operator (**The roster**,
above), naming the gating sum, its per-worktree counts and, when it differs, the full-branch sum.
Record both with `flow record pass -round <round>` for this round,
alongside the agents-ran/why/diff-path lines the fix pass records.

**The docs-only guard runs again beside that cap check**, `check-panel-docs-only.sh <worktree>
<merge-base>`, per design.md's `fix-rounds-reclassify`. A branch that stays docs-only keeps the
reduced roster, and `primary` re-runs on its delta as above. A branch the fix round made no
longer docs-only — exit 1 or 2 where pass 1 saw exit 0 — dispatches, in this round, every
resolved slot not yet dispatched this run, each reading the whole `final-review.diff` under the
no-last-reviewed-sha rule above; Bugbot, Mutation and Security among them take their pass-1 shape, throwaway
worktree included. Record which slots joined this way and the path the guard printed.

Handoff still requires **zero open findings at any severity** from every agent that has run, and
no stale result — where **a slot's clean result is stale when the rule above required that slot to
re-run and it has not, or when any commit or working-tree change to source landed after that slot's
last read, from any stage — `flow.verify` included**. An unrecorded edit after the panel closes is
stale by definition, not only one a fix round produced. A fix against which a slot raised no
finding leaves that slot's result current: the round's own mutation-proof (below) covers what the
fix changed.

Union all **open** findings, dedupe by **defect identity — file:line + theme.** *File:line* is the
finding's own recorded location, taken verbatim from the findings table. *Theme* is the finding's
one-sentence Note column, reduced to its own defect noun phrase — the shortest phrase naming what is
wrong, severity words and slot names stripped out.

**Before dispatching the fix subagent**, the parent runs this itself, never a subagent:

```bash
check-panel-reproducers.sh <worktree> <change>
```

Exit 0 proceeds. Exit 1 covers two classes: a **missing or malformed field** is added before
dispatch; a **rejected reproducer shape** (a shell metacharacter, an absolute path, a `..` segment,
a leading `-`, a URL, a NUL byte) is a **refusal** — the line is recorded **unverifiable** and put to
the operator, never silently rewritten. Exit 2 stops the run.

**For each open finding whose record carries a runnable `finding-reproducer:` command**, the
parent runs it itself, never a subagent — every finding's run and every throwaway worktree
removal in one Bash call, each run followed by `; echo "F<n>: exit $?"` so every exit code stays
readable:

```bash
run-reproducer.sh <worktree> "<the finding's finding-reproducer: text>"
```

Read its exit code: **0** dispatches the finding; **1** bounces it once, back to the raising slot,
carrying the reproducer's passing output; **2** is a refusal — recorded **unverifiable**, put to the
operator; **3** is a timeout or a detached survivor — recorded **unverifiable**, put to the
operator, with a surviving pid named when the script names one; **4** stops this finding's dispatch
decision entirely — a finding recorded `none — <reason>` is dispatched without a run.

A slot that supplies nothing for a finding has not supplied a legal exemption: record that omission
as its own open finding.

**Once the fix subagent reports, re-run every dispatched finding's reproducer** under the same
constraints and require it now to exit **0**. **The parent runs these re-runs itself, in its own
Bash calls — never a "verify fixes" reader or any other subagent** (**Dispatch sites — the
parent's closed list**, `skills/flow/implement.md`). **The flip alone does not close a finding — the fix's
diff must also touch at least one path the finding named, with a non-comment, non-whitespace
change.** A fix that does not is not a fix: the finding stays open and goes to the operator through
the handback below.

A finding meeting both conditions is recorded closed there and then:

```bash
flow record status -change <name> -ref F<n> -status fixed
```

**The parent records it, never the fix subagent.** **Record every verdict this turn reached in one call,
never deferred to the round's end** — the reproducer re-runs and the fix diff are read in one
call, each finding judged, then every `status fixed` recorded together, so an aborted round
still leaves every already-verified finding closed. **A finding failing
either condition is left untouched** on `open`, for the handback below. This walk follows
**Read discipline** (`skills/flow/implement.md`): the specific hunks a finding names, never the <!-- refs-guard:allow -->
whole fix diff.

### The fix round mutation-proves what it changed

**Every executable behaviour the fix changed is mutation-proved, not only the test cases the round
adds.** The fix subagent performs the proof and reports it, per the MUTATION PROOF paragraph its
dispatch carries; the parent runs no build of its own here. A survivor the fix subagent cannot
judge real or equivalent goes to the operator through the same handback the section already names.

**Record each one with one `flow record mutation -change <name> -round <round> -path <path>
-mutated <what> -test <test>` call per line, transcribed from the fix subagent's report — the
exemption form records `-mutated none -test <reason>`; the lines the calls write are:**

```text
fix-mutation: <path> — <what was mutated> — <the test that failed>
fix-mutation: <path> — none — <reason>
fix-mutations-total: <n>
```

**The rendered record carries these lines in its pass-log section, one round's count after that
round's lines, and never inside the marker block.**

**No line anywhere in the panel record may carry the literal label `finding-status:`,
`findings-total:`, or `finding-reproducer:` outside its own marker use.** Write around it: paraphrase
the label, or break it with a non-word character.

**The parent checks the reported list against the fix diff before the round can close, reading the
`fix-mutation:` lines and walking the diff itself — never a "mutation re-verify" subagent or any
other delegate.** Walk every
hunk of the fix diff with a non-comment, non-whitespace change: each one is either covered by a
reported line, or is not an executable behaviour at all. A hunk that removes or weakens a test or an
assertion states in the record what it used to cover and names what still covers that same behaviour
now — checked by running the named covering test against the **pre-fix** code and confirming it
fails. A hunk whose path is draw/geometry code is held to the **PIXEL PROBE** paragraph
beside the mutation lines: the fix's report names the probe assertion it landed against the
actual rendered pixels or geometry, and a fix commit for this class of bug that landed no such
probe is rejected at the fix step — the round does not close and the finding stays open, for the
handback below — rather than the regression being discovered next round (KAN-496). **The same
walk holds the fix subagent to its PLAN FIELDS obligation:** a hunk that adds a
test case, adds a file, or changes what a task's `**Baseline:**` counts, whose task's `**Tests:**`,
`**Baseline:**` or `**Files:**` field in `<changeRoot>/tasks.md` does not reflect it, does not close
the round; it goes to the handback.

Beside the reproducer re-runs above, the round close re-runs the task-field guard mechanically:
`check-task-commit-fields.sh <worktree> <task-id> <task-sha> "" <canonical-worktree>
<name>` for every task a fixup folded into — `<task-id>` from that commit's `Task-Id:` trailer,
`<task-sha>` the folded commit as it now stands, the remaining arguments resolved the way
`skills/flow/implement.md`'s task-close step resolves them. A non-zero exit does not close the
round; it goes to the handback. This catches an undeclared file the fixup added and a declared
test it removed or renamed, read post-autosquash; a test added to the commit without a
`**Tests:**` declaration and a stale `**Baseline:**` count remain the walk's judgment until the
companion checker (KAN-511) lands.

This binds the fix round every run — the obligation is the round's, not a slot's, so a run where
neither Bugbot nor Mutation is in the resolved roster or added this run is exactly where the round's own proof
is the only mutation reasoning that happens at all.

**Rebuild the dispatch context bundle before dispatching the fix subagent**, same as above,
overwriting the same path. Report the script's stderr line (`bundle unchanged — reusing …` or
`bundle rebuilt — …`) as part of this round's own reporting.

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

**Every fix subagent's dispatch prompt also carries the PLAN FIELDS paragraph**:

> **PLAN FIELDS:** your fix owns its plan record, as part of the fix itself: when it adds a test
> case or changes what tests a task names, update that task's `**Tests:**` field; when it changes
> what a task's `**Baseline:**` counts or `**Files:**` paths declare — a test case added, a file
> created — update those fields too. All of it lands in the worktree's
> `<project>/spectre/changes/<name>/tasks.md` in this same pass — never left for a reviewer to
> catch next round (kan-454, KAN-459). Edit them; do not stage or commit them — the plan record is
> a planning path and is committed later by the pipeline, never in a fixup.

**Every fix subagent's dispatch prompt also carries the ROUND SCOPE paragraph**:

> **ROUND SCOPE:** your round's output is the diff and the report, nothing else. Do not rewrite
> `proposal.md` or `design.md`: unaffected sections are never restated, and a decision your fix
> genuinely overturns is superseded by one appended entry under `## Decisions`, never a per-round
> rewrite of the file. Deferral rationale is written nowhere — a Minor is deferred with its
> one-clause reason in the store. Your report names what you fixed, the behaviours you changed,
> and the task commits your fixups folded into, then stops.

**Every fix subagent's dispatch prompt also carries the FOREGROUND BUILDS paragraph**:

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

**Every fix subagent's dispatch prompt also carries the TOOLS paragraph**:

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

**Every fix subagent's dispatch prompt also carries the NO DELEGATION paragraph**:

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the parent's closed list (**Dispatch sites — the
> parent's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

**Every fix subagent's dispatch prompt also carries the MODEL HANDSHAKE paragraph**:

> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.

Dispatched on `DEFAULT_MODEL` (below); the dispatcher compares that line against it and applies
**The handshake** (`skills/flow/implement.md`, **The parent orchestrates directly**), unchanged: a first
mismatch is a fallback plus one retry under `<round>-fix-retry`; a second is a fallback plus
`## Question`.

**Every fix subagent's dispatch prompt also carries the TARGETED TESTS paragraph**:

> **TARGETED TESTS:** Run only the tests this task's `**Tests:**` field names, through the build
> tool's own selector — `--tests '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t
> '<name>'` for vitest — once for RED, once for GREEN, and again only after a source edit. Never
> run the module or repository suite mid-task: the full `## test` list runs once per worktree at
> the last bundle, and again in `flow.verify`. Pipe a test run's output through `tail` so a green
> run costs lines of context, not a build log.

**Every fix subagent's dispatch prompt also carries the MUTATION PROOF paragraph**:

> **MUTATION PROOF:** every executable behaviour your fix changed is mutation-proved before you
> end your turn — not only the test cases this round adds. Mutate the mechanism: revert it in a
> scratch tree, or flip the single value it turns on — but before the tests run, confirm the edit
> landed: the target changed where you intended, not nowhere and not somewhere else. An edit that
> never applied is a refusal, not a surviving mutant — redo it with a working mechanism; it never
> buys a test. Then confirm an existing test fails, and restore.
> `mutate-and-verify.sh` mechanizes backup, apply, run, report and restore
> for a mutation expressed as a patch file against one or more test harnesses; which mechanism to
> mutate is your judgment, not the script's. Each mutation alters one mechanism — where a single
> revert would also change state a second check reads, split it into surgical mutations, one per
> mechanism. A mutation no test catches is a surviving mutant: add the test that catches it before
> your turn ends. Record one `fix-mutation:` line per behaviour in your report, plus a
> `fix-mutations-total:` count, in the shape the review-panel contract's fenced block gives. Where
> you cannot judge whether a survivor is real or an equivalent mutant, say so in the report rather
> than deciding it yourself.

**Every fix subagent's dispatch prompt also carries the PIXEL PROBE paragraph**:

> **PIXEL PROBE:** every fix you land to draw/geometry code — code that computes what is
> drawn: extents, bounds, gridlines, offsets, paths, positions, sizes — carries a probe assertion
> against the actual rendered pixels or geometry, in the style the reviewers' reproducers
> already use: drive the real draw path and assert on the value it renders or computes, never on
> a hand-derived copy of it. A fix commit for this class of bug that lands no such probe is
> rejected at the fix step — the finding stays open and the round does not close — so the
> regression is caught this round, not discovered by the next one. Where pixels cannot be
> rendered in this environment, the probe asserts on the geometry the draw path actually
> computes — the same numbers the renderer will paint — and your report names why the pixel
> output itself was not asserted.

**Every fix subagent's dispatch prompt also carries the REPORT FILE paragraph**:

> **REPORT FILE:** write your report to
> `<abs-worktree>/.superpowers/sdd/panel-fix-report-<round>.md` — a chunked round's chunk `<n>`
> appends its own suffix, `panel-fix-report-<round>-<n>.md` — as your **last** act — after
> the rebase and your final test run — naming each finding you addressed, the executable
> behaviours your fix changed, the `fix-mutation:` and `fix-mutations-total:` lines this round's
> contract requires, and the task commit each fixup folded into. The dispatcher waits
> on that file's presence.

Give the surviving findings to fix subagents in **chunks of at most 10 findings**. The round's
findings are split into sequential chunks of at most 10 — the cap that keeps one dispatch from
re-reading the whole branch across every finding — each chunk one panel-fix dispatch carrying its chunk of
the combined list. Never one dispatch per reviewer, per slot, or per finding — that split
fragments one diff into competing fixups against the same worktree. Chunks run in order, each
dispatch awaited in the foreground before the next chunk begins and before the round's reproducer
re-runs begin, and no fix subagent is left in flight when the turn ends. The round's first
chunk's `-key` is exactly `panel-fix-<round>`; each further chunk appends `-<n>`
(`panel-fix-<round>-2`, `-3`, …), contiguous from 2; the handshake retry suffixes `-retry` onto
its own chunk's key (`panel-fix-<round>[|-<n>]-retry`), and any other key shape is a violation
whatever the dispatch count. A round may not chunk freely: its chunk count is bounded by
`ceil(findings raised in earlier rounds / 10)` — a bound, not a target, so a well-formed
per-finding sequence stays caught. Before recording each `dispatch begin`, confirm that key has
no panel-fix begin already recorded — a second begin under a fresh key is over-dispatching even
when every key is well-formed. `check-panel-fix-single-dispatch.sh` holds every run's panel close
to exactly this shape (**Before closing the stage**, below). Inline
(`skills/flow/implement.md`'s **Inline — the parent implements**), the parent applies the fix
itself under the same paragraphs, dispatching no subagent, and records the pass with `-role
panel-fix -agent-id inline`. Where a finding is
confirmed as a real defect, the fix subagent invokes **superpowers:systematic-debugging** before
writing its fix. **Dispatch it on `DEFAULT_MODEL`** (design.md's `model-default-sonnet`). **On
`IMPLEMENTER_MODEL_TOGGLE` `dynamic` with an `sdd` decision, dispatch it instead on the decision's
`implementer` object** — the fixer's own model and effort, `subagent_type:
flow-<effort>` with that `model` passed as the Agent tool's own `model` parameter, and
`-model`/`-effort` below carry that pair. Record
every pass with `flow record pass -round <round>`: which agents ran, why,
the diff path they read, and — when this pass bounced any finding — each bounced finding's defect
identity together with the reproducer output it carried back.

**The fix subagent's own dispatch is recorded too, with `-role panel-fix`:**

```bash
flow record dispatch begin -change <name> -role panel-fix -model <m> -effort <e> \
  -key panel-fix-<round>[-<chunk>] -agent-id <id> -session-token mf-<literal-token> -started-at <ts>
flow record dispatch end -change <name> -key panel-fix-<round>[-<chunk>] \
  -session-token mf-<literal-token> -commit <partner-task-sha> -outcome completed -ended-at <ts> \
  -agent-id <id>
```

`-commit` is the task commit the fixup was folded into.

A Minor either fixed or deferred blocks nothing; a Minor left `open` blocks exactly as a Critical
does. When fix rounds do not converge,
the run hands back to the operator, one finding at a time:

> **`<location>` — <the finding, in one line>. The fix round did not resolve it.**
> - **Take another round on it** *(default, recommended)*
> - **Withdraw it — I'll give the reason** — the reason is recorded on the finding's marker line
> - **Stop the run and hand it back to me**

Only that answer records `withdrawn`, and only with the reason the operator gives.

**Before closing the stage**, the parent runs both close guards itself, never a subagent:

```bash
check-panel-findings-closed.sh <worktree> <change>
```

Exit 0 proceeds to the stage close below. Exit 1 means a finding still reads `open` in the store —
return to the handback loop above for it. Exit 2 stops the run.

Beside it, run

```bash
check-panel-fix-single-dispatch.sh <worktree> <change> <session-token>
```

— the token this run stamped on its own dispatches. Exit 0 proceeds to the stage close below.
Exit 1 names every violation of the chunked fix-dispatch contract above — an over-bound chunk
count, non-contiguous chunks, a chunked round with no bare key, a per-chunk count or retry
violation, a key outside the canonical shape — and is a handback `## Question`:

> **The fix round(s) broke the chunked fix-dispatch shape:** <the guard's violation lines>
> - **Continue — the violation stays recorded in this run's output** *(default, recommended)* —
>   the findings may already be verified closed, and the round's work is real
> - **Stop the run**

Exit 2 stops the run — a close the guard cannot answer for is not a clean close.

```bash
flow stage end -command '/flow' -stage flow.review-panel -outcome completed <name>
```

Once this stage ends clean — zero open findings at any severity, no stale result as **Panel
re-runs** defines it — continue into
`skills/flow/verify-and-handoff.md`.
