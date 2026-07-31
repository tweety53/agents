# Fix wave — kan-17-finish-gate-jira-and-commit-hygiene

One consolidated wave over the whole panel list. Every row in
`.superpowers/sdd/final-review-panel.md` is now `fixed` (28) or `withdrawn` (2, both withdrawn by
the panel itself before this wave). **Nothing was withdrawn by this wave** — every finding was
judged real and every one was fixed. The two disagreements I have are recorded at the end as
disagreements with a *rationale*, not as refusals: both were fixed anyway, differently from the
shape the finding proposed, and the difference is stated.

## Per finding

### C1 — the findings parser fails open (`scripts/check-unfinished-work.sh`)

Replaced the positional regex with a real table parser:

- The findings table is located by its **header row** (`Slot | Severity | Location | Status | Note`,
  matched case-insensitively on all five names). A reordered table is not recognised, and a panel
  record with no recognisable findings table reports outstanding.
- Rows are **split into cells**, honouring `\|` as an escaped pipe. A row that does not read as
  exactly five cells is reported as unparseable, never skipped.
- The status cell is compared **for equality, case-insensitively**, against `open` / `fixed` /
  `withdrawn`. Anything else — blank, capitalised, commented — counts as outstanding.
- A `withdrawn` row with an empty Note counts as outstanding (this is also I5's guard half).

Column order pinned normatively in
`openspec/changes/.../specs/myflow-review-panel-economics/spec.md` (a SHALL, plus the status cell's
position, plus the escaping rule and three new scenarios), and called load-bearing in
`skills/myflow-do/SKILL.md` with the reason (a script counts that cell).

**Verification.** Harness cases 4a–4i cover all four demonstrated shapes plus a blank cell, a
table-less record, and both directions of withdrawal. Two cases assert the *opposite* direction so a
parser that refused everything could not pass: an escaped pipe in a note still reaches `CLEAR`, and
a five-column table that is not the findings table (the roster) is not counted. Running the new
parser over this change's own panel record immediately found four rows with genuinely unescaped
pipes (notes about `\|\| true` and `open\|fixed\|withdrawn`); they are now escaped, which is the
guard working rather than a false positive.

### C2 — a missing primary plan file reports CLEAR

The primary `openspec/changes/<name>/tasks.md` is now tracked separately from the optional fix
sub-change plans and reports `no plan at <path>`. Harness: a missing change directory, and no
`openspec/` tree at all.

### I1 — the two-commit sequence breaks in four ways

All four fixed by one change of shape, in `pipeline.md`'s **Git boundaries** (canonical) with both
skills pointing at it:

```bash
git -C <wt> reset -q -- openspec/ docs/manual-test/ docs/superpowers/ \
  && git -C <wt> add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/' \
  && { git -C <wt> diff --cached --quiet || git -C <wt> commit -m "..."; } \
  && git -C <wt> add -A \
  && { git -C <wt> diff --cached --quiet || git -C <wt> commit -m "chore(...): ..."; }
```

**Verification — each of the four reproduced in a scratch repository, before and after:**

| Case | Before | After |
|---|---|---|
| Only planning paths changed | `git commit` exits 1, "no changes added to commit" | chain rc=0, implementation commit skipped |
| Only implementation changed | second commit exits 1 | chain rc=0, planning commit skipped |
| Re-run after a failed push | first `git commit` exits 1 on the wrong step | chain rc=0, no-op, log unchanged |
| Hook rejects the first commit | sole commit titled `chore(x): plan, test guide and session records` **contains `src/a.kt`** | chain rc=1, stops, no commit made |

The hook case needed the realistic trigger to reproduce: a hook that fails on the first run and
passes after — the `pre-commit` framework's auto-format behaviour. A hook that fails
deterministically on content fails both commits and produces nothing, which is not the reported
defect.

**One correction to the proposed fix shape.** The Principles slot proposed `set -e` plus the `||`
guard. I verified `set -e` is not a reliable stopper here: bash 3.2 (the platform shell, and this
repo's stated floor) ignores `set -e` inside a subshell whose parent has errexit off, and this block
is run through an agent's shell whose state the skill does not control. Reproduced: the same
`( set -e; … )` block ran on past the rejected commit. The sequence is a single `&&` chain instead,
which needs no shell state at all. The contract says so, and says to run it as one command.

### I2 — a symlinked excluded path is fatal

Reproduced: `git add -A -- . ':(exclude)docs/superpowers/'` against a tracked
`docs/superpowers -> ../real-sp` exits 128 and stages nothing. **Decided and documented as
stop-and-report**, in `pipeline.md`'s Git boundaries, with a guardrail in `myflow-finish/SKILL.md`
against the one workaround: a bare `git add -A` is the only way to stage past it and it puts the
planning artifacts into the implementation commit — the outcome the whole split exists to prevent.
The fix belongs in the repository, by making the path a real directory.

### I3 — `LEFTOVER:` never blocks the `FINISHED` write

Chose **the verdict gates the write**, not a new durable field.

`LEFTOVER:` and a missing verdict both leave the change at `IN_PROGRESS`, where `/myflow-status`
still lists it and `/myflow-finish` still acts on it — so the state the change already carries *is*
the durable record. A state-file field was considered and rejected: `design.md`'s non-goals exclude
"any change to … the state file's schema", and inventing a field to carry a fact the state itself
carries would be the weaker fix.

Run 2 re-entrancy is now stated, because that is what makes the gate terminate rather than wedge:
every step is remove-or-move *if present*, an already-gone artifact is success, and an
already-archived change directory means the sync-and-archive step is skipped rather than repeated.
An absent verification script falls back to checking the same rows by hand and saying so — the same
precedent `check-finish-preflight.sh` already has — so a repository without the script is not
wedged either. A second handoff shape ("Cleanup incomplete — not finished") names what remains and
points back at `/myflow-finish`.

Spec amended (`specs/myflow-finish-cleanup/spec.md`): the blocking rule, the re-entrancy rule, the
both-directions coupling rule, and three new scenarios.

### I4 — no containment on the PR-controlled `<change-name>`

Both guards now apply `preserve-session-records.sh`'s Protection 1 allowlist, character for
character, **before any path is built**, exiting 2 with no verdict line.

**On reuse vs. a shared helper.** The finding says "reuse that allowlist". I copied the `case` block
rather than sourcing a shared file, and the copies say why: these guards are single-file by design
and get copied into projects one at a time — the finish contract has an explicit "when the script is
absent" path — so a sourced helper would produce a guard that is present but unrunnable, which is a
worse failure than six duplicated lines. To stop the copies drifting, **both harnesses assert the
same five rejected shapes**, and both also demonstrate the traversal end to end rather than only
asserting the error message.

### I5 — `withdrawn` is an unguarded escape hatch

Both halves closed. The guard counts a `withdrawn` row with an empty Note as outstanding (harness
4i). The contract half is now operationalised: the handback is an actual named-options prompt, one
finding at a time, and `myflow-do/SKILL.md` states that no fix subagent may write `withdrawn` at
all — a fixer marks `fixed` or leaves `open`, and a disagreement goes in the note and the report,
never in the status. (This wave followed that rule: nothing here was withdrawn.)

### I6 — the verifier's rows are unlinked from the registry

Made explicit **and checkable in both directions**. `check-cleanup-complete.sh` carries one
`registry-row-checked:` / `registry-row-not-checked:` marker per registry row, with a reason on each
not-checked line. Harness case 14 reads the real registry out of `pipeline.md` and the markers out
of the guard and fails when they disagree either way.

**Verification — mutation-tested both directions:** adding a `Scratch cache` row to the registry
fails the suite ("registry row(s) the guard declares nothing about: Scratch cache"); adding a
declaration for a row the registry lacks fails it too ("declares row(s) absent from the registry:
Imaginary row"). Both restored, suite green.

### I7 — the unrecognised-status ask contradicts the never-blocking guardrail

Resolved by **carving it out explicitly and bounding it**, in both files. `jira-integration.md` now
gives the ask a named-options shape and states it is the single carve-out from **Never blocking**,
with limits: asked once per run, never repeated, never retried, reached only when an unrecognised
status was actually observed, and **only an explicit yes transitions** — No, silence, a non-answer,
or a session that cannot ask degrades to the same one skipped-with-reason line, and the proposal is
untouched either way. `myflow-start/SKILL.md`'s guardrail names the carve-out and points at it.

### I8 — the gate's issue-creation has no failure contract

Creation joins the enumerated failure paths under **Never blocking**, and has its own paragraph
under the labels rule listing what refuses it (auth, permission, unknown project key, disallowed
label, missing required field, no tooling at all). Both call sites — `pipeline.md`'s course table and
`myflow-finish/SKILL.md` §1.0 — state that a failed filing is one line and the run still continues,
that the outstanding list still reaches the planning commit and the handoff, and that a failed
filing is never silently converted into a different answer.

### I9 — the change violates its own planning-gate rule

`myflow-do/SKILL.md` §3's "Ask which" is now two named options with a marked recommendation
(appending, because most fixes are corrections within existing scope) and each option states what it
does. The text also says why the rule applies here: the planning-gate capability governs choices
"wherever a `/myflow-*` command asks", not only in `/myflow-start`.

### M1 — restated pathspec reasoning

`myflow-do/SKILL.md` now points at **Git boundaries** and keeps only what is specific to it: whose
staging it retracts, and why `git reset` rather than `git rm --cached` or `git restore --staged`.

### M2 — two verdict vocabularies

Fixed by **stating the reason in both headers** rather than unifying. The reason is real and was
already half-written in one of them: the two guards answer opposite questions about absence — a
missing file counts *against* the change in the run-1 gate and *is the answer* in the run-2
verifier — so shared words would carry that inversion between them. The third vocabulary
(`RUN1`/`RUN2`/`REFUSE`) is noted as differing for a further reason again: it selects a procedure.

### M3 — self-citation in `engineering-principles.md`

Back to one sentence, with a line recording why those three group names carry no path: they are the
next three headings of the same file.

### M4 — `count_unticked`'s blanket `|| true`

Exit 0 and 1 return the count; exit ≥ 2 returns non-zero and the guard refuses with exit 2 and no
verdict line, naming the file. An unreadable file is an honest unknown, not zero unticked boxes.
Harness case 8d uses a mode-000 file: rc=2, empty stdout, `check-unfinished-work: cannot read …` on
stderr.

### M5 — the `## Known incomplete` scraper is not fence-aware

Made fence-aware, and the header now states the **rule that decides** where fence tracking is paid
for: wherever ignoring fences fails *open*. It does here — a fenced example carrying the heading and
`None.` would clear the signal for a guide that never wrote the section — and it does not for the
checklist pattern or the findings-table parser, both of which err toward outstanding. Cases 8b (a
fenced example is not the section) and 8c (a fenced item *inside* the section is still content).

### M6 — the untested refusal path

Fixed by the first of the two options offered. Harness case 13 puts a stub `git` earlier on `PATH`
that fails `worktree list` and `exec`s the real binary for everything else — so `rev-parse
--git-dir` still answers honestly — and asserts the refusal shape (non-zero, empty stdout, named
stderr).

### M7 — fix sub-change glob shapes

Fixed by removing the dependence on the layout rather than by adding a fourth glob. Every
`tasks.md` at any depth under the change's own directory, and under any sibling `<name>-fix-*`, is
read. Cases: a doubly-nested fix-of-a-fix, a sibling fix-of-a-fix, and (the other direction)
another change's unchecked plan is still not counted.

### M8 — the three-course prompt marks no recommendation

Fixed rather than withdrawn: there is an honest default. **Stop — I'll finish it first** is
recommended and listed first, because the gate only fires when something genuinely is unfinished and
**Continue** is the only course that reaches an irreversible step. The reasoning is written beside
the prompt, including that Continue exists for work the operator deliberately deferred.

## Where I disagree

Neither of these is a refusal — both findings were fixed — but the disagreement is on record.

1. **The Principles slot's proposed fix shape for I1 (`set -e` plus a `||` guard) does not work on
   this platform.** Verified above. Using it would have left the hook-rejection half of the finding
   open while looking fixed, which is worse than not fixing it. The `&&` chain is used instead.

2. **I4's "reuse that allowlist" is implemented as a duplicated block, not a shared file.** I think a
   sourced helper would be the wrong call for scripts that are explicitly allowed to be absent and
   are copied one at a time; the coupling is enforced by both harnesses asserting the same rejected
   shapes instead. If the panel meant literal code sharing, this is a deliberate deviation with the
   reason recorded in both guards.

## Round 2 — one Critical the round-1 rewrite introduced

The scoped re-review verdicted all 22 findings addressed and reproduced the `set -e` claim
independently. It raised one new Critical **in the parser round 1 wrote**, which is the honest
outcome: the rewrite closed four evasions and opened a fifth, larger one.

### The defect

Signal three's state machine ended the table at any line not starting with `|` after trimming
(`if (line !~ /^\|/) { in_table = 0; next }`), and only a fresh header row could restart it. A row
missing only its **leading** pipe is valid GFM that renders identically to the piped form, so that
shape hid the row *and every row below it*, with no `unparseable` and no `unrecognised` count — a
single character removing an unbounded number of findings, defeating exactly the mechanism C1
exists to protect.

**Reproduced before the fix**, both shapes, each printing `CLEAR:` over an open Critical:

- a `fixed` row, then an `open` Critical missing its leading pipe, then a further well-formed `open`
  row — all of it dropped;
- a blank line inserted before an `open` row — dropped by the same rule.

### The fix

1. **Boundary pipes are optional.** `cells_of()` splits on unescaped pipes and drops a leading
   and/or trailing empty field when present, so `| a | b |`, `a | b |`, `| a | b` and `a | b` all
   yield the same cells — for the header, the delimiter and every row. Cell indices are now 1-5
   directly rather than 2-6 of a seven-field split.
2. **What ends the table is a closed, stated list:** a blank line, an ATX heading, or end of file.
   **Never "a line the parser did not recognise"** — that rule *was* the defect. Any line inside a
   table that does not read as five cells is counted as unparseable, and the rows after it are still
   read. The header comment states this exhaustively, with the deliberate consequence spelled out: a
   block abutting the table with no blank line before it reads as an unparseable row and reports
   outstanding, which is the safe direction.
3. **A findings row outside any findings table is still reported.** Ending the table at a blank line
   is what the format says, but alone it would leave a smaller version of the same defect — one
   blank line detaches a row. So any line anywhere in the record that reads as exactly five cells
   whose fourth is `open`, `fixed` or `withdrawn`, and is not inside a recognised table, is counted
   and named. The roster table is unaffected: its fourth cell holds `yes`, never a status word —
   asserted as its own case.

### Verification

Harness cases 4j-4p; that suite is now 88 assertions (was 75):

| Case | Asserts |
|---|---|
| 4j | a row missing its leading pipe is OUTSTANDING, and **`2 open finding(s)`** — the count, so a guard noticing only the first row cannot pass |
| 4k | a row missing its *trailing* pipe is read, not rejected — the false-positive direction |
| 4l | a table written with no boundary pipes at all is still a findings table |
| 4m | a stray prose line inside the table does not end it: the row after it is still counted **and** the stray line is reported unparseable |
| 4n | a row detached by a blank line is reported as outside any findings table |
| 4o | the same for a row parked under a later heading |
| 4p | the roster and convergence tables are not read as detached findings |

The real panel record still parses: 1 findings table, 0 unparseable, **0 open**, 29 `fixed`,
2 `withdrawn` — the new finding is recorded there as a row, `fixed`, with a pass-2 note.

### Re-check of the other two signal parsers — asked for, and done

**Signal four (`## Known incomplete`) does not have this bug.** Its section ends only at a `^## `
heading outside a fence; every other line is *collected as content*, so an unrecognised line cannot
terminate it. Verified empirically across ten shapes: prose that is not a bullet, a table, an `h3`
sub-heading, a later `h1`, a thematic break, blank-only content, a heading with a suffix, and the
`h2` control. Every deviation errs toward OUTSTANDING; none hides content.

**Signals one and two have no state machine at all** — `grep -c` on an anchored pattern, plus a
`find` and a `case` filter. Verified that an unticked box after a fenced block, and an unchecked
plan item after odd content, both still count.

### One residual found during that re-check — reported, not fixed

The section scraper's **fence toggle has no nesting or indentation rule**: any line whose first
non-space characters are ``` or ~~~ flips the state. An *odd* number of fence-looking lines can
therefore leave `in_fence` at 0 where it should be 1, and a fenced *example* of the section is then
read as the real one.

**Reproduced.** A guide with no real `## Known incomplete` section, containing (a) a fenced block
that itself shows an indented ``` marker — which this repository's own docs write — and (b) an
unclosed fenced example of the heading followed by `None.`, reports **`CLEAR:`**. The control (the
same guide without the nested-fence block) correctly reports "no '## Known incomplete' section".
The unclosed fence is load-bearing: with the example closed, its closing fence line is collected as
content and the verdict flips back to OUTSTANDING, which is why both ingredients are needed.

**Not fixed here, deliberately — this is a decision for the coordinator.** It is a different defect
class from the one I was sent to fix (not "an unrecognised line terminates the state machine" but
"the fence toggle is not a fence model"), and the instruction was to fix that one and nothing else.
More importantly, a *partial* fix would be worse than none: matching CommonMark's indent rule
(≤ 3 spaces) kills the reproduction above but leaves the marker-**length** rule open — a ```` fence
documenting a ``` fence unbalances the toggle just the same — and shipping a fix that looks complete
while the class stays open is the exact failure this change exists to end. A correct fix is a real
fence model (opening marker character, length and indentation, with the closing fence required to
match), which belongs in front of the panel rather than inside a scoped round.

## For the controller

Four of this change's **design decisions** now describe behaviour the fix wave changed, and
`design.md` was deliberately not edited — the brief restricted `openspec/` edits to the two named
delta specs, and recording a fix in `proposal.md`/`design.md` is `/myflow-do` §3's job, not a fix
subagent's. The decisions that moved:

- `unfinished-work-prompt` — the three courses now carry a marked recommendation (**Stop**).
- `cleanup-verification-script` — a `LEFTOVER:` verdict now blocks the `FINISHED` write, and run 2's
  re-entrancy is part of the contract.
- `jira-unknown-status-ask` — the ask is now an explicit, bounded carve-out from **Never blocking**.
- `findings-table-and-bar` — the table's column order and pipe escaping are now normative, and
  `withdrawn` with an empty note does not close a finding.

## Not addressed — out of scope for this wave

`docs/manual-test/kan-17-…md`'s `## Known incomplete` section still records two items, neither of
them a panel finding: the third family of unchecked cross-references at
`skills/myflow-contracts/state-file.md:45` (documented as needing a structural decision), and the
two surfaces with no automated coverage (Jira MCP calls, operator prompts). The guide's 20 unticked
boxes are the operator's manual test pass, which has not happened yet. `check-unfinished-work.sh`
therefore still reports `OUTSTANDING` for those two signals — and reports **no open findings**,
which is what this wave was asked to reach.
