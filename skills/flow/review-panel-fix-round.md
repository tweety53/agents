# Review panel — fix rounds

Loaded during `flow.review-panel` by the load directive of **Panel re-runs**
(`skills/flow/review-panel.md`), only once a round has recorded a Critical or Important finding or a close guard sends the run back
to the handback loop, or read by section where a citer outside the panel stage names it.
Every section name below without a path is a section of this file or of
`skills/flow/review-panel.md`; the fix subagent's dispatch paragraphs and its dispatch record stay
there.

## Panel re-runs

**Every round this stage dispatches after pass 1 — each fix-round re-run — opens by running **Check base movement first** again: the same
per-worktree `sync-panel-base.sh` call, with that section's
verdicts, automatic no-overlap rebase, overlap prompt and conflict handling unchanged — a `MOVED`
with no overlap rebases unasked at a round boundary just as at entry.** **Continue** at a round
boundary proceeds into the round's own remaining steps — the citation pre-check that follows the
entry check is an entry step, and the round does not re-run it.

**Pass 1 runs the roster **The docs-only reduction** or **The late-fix reduction** chose — the
resolved roster, `primary` alone on a docs-only branch, or `primary` alone on the late-fix delta
where the reduction's trigger holds — plus every slot the operator named at this stage's start
that it did not already carry.** Only re-runs after a fix are scoped. Record
`FIX_BASE` — the branch tip the fix round starts from, per worktree — commit the fix, then write
`<abs-worktree>/.superpowers/sdd/fix-round-N.diff` and its touched list with
`write-panel-diff.sh fix-round <N> <abs-worktree> <worktree> <fix-base> [<worktree> <fix-base>…]`,
whose exits read as the pass-1 write's (**The docs-only reduction**, `skills/flow/review-panel.md`).

**A branch the remote already holds takes the fix as one new commit on top, never a rewrite.**
This is the normal case: the branch is pushed with every commit (**Branch backup**,
`skills/flow-contracts/git-boundaries.md`), so the fix stages the changed paths
(`git add -- <the changed paths>` — a pathspec commit reads tracked paths only) and is a plain
`git commit -m ... -- <the changed paths>` at the tip — the pathspec-scoped default (**Branch backup**, `skills/flow-contracts/git-boundaries.md`),
so it carries only the paths the finding named, whatever else the index holds — pushed
plain like any other commit, and every downstream commit keeps its sha.
**Rewrite-based folding is for unpushed history only**: the fixup targets the **original** task
commit (`<task-sha>`), and the autosquash folds it in immediately, before anything pushes — one
call, with the changed paths:

```bash
fold-fixup.sh <worktree> <task-sha> <changeRoot>/tasks.md <the changed paths>
```

Its header (`<agents repo>/scripts/fold-fixup.sh`) is canonical for the sequence it runs. The
route's own diff is `git diff "$FIX_BASE"..HEAD`, read once the fold has landed. Exit 0 → folded
(`FOLDED:`), or, where the fold emptied its target commit, that commit dropped (`DROPPED:`) — the
plan's task entry stays, the record of the mistake the branch no longer carries.
`git diff "$FIX_BASE"..HEAD` then reads the restoration itself, and the round's reproducer
re-runs and diff check close on it unchanged. Exit 1 → a `guard-autosquash.sh` refusal: the round stops before anything
builds on the rewritten history; exit 2 → it cannot answer: report its stderr and stop the round;
exit 3 → a conflict between two of the branch's own commits is left in progress (`CONFLICT:`):
resolve it by hand, keeping both sides, run `git rebase --continue`, then
`fold-fixup.sh --finish <worktree> <task-sha> <changeRoot>/tasks.md`, whose exits read as the
first call's.

**A clean `git rebase --autosquash` is not evidence the fix survived it.** Where the fixup and the
commit it folds into touch nearby lines, git's 3-way auto-merge can resolve in favour of the
pre-fix side — it exits 0, prints no conflict marker, and leaves no `fixup!` commit behind. The
reproducer rerun and diff check below (**Once the fix subagent reports…**) are what catch this;
they must run against the post-rebase file content, never be satisfied by the fixup commit's
presence or the rebase's own exit code.

**Which slots re-run, and on what, follows from the findings the round fixed and the severities
it raised — never from a mode table, a trigger list, or a round count.**

**A Critical is confirmed against a primary source before any fix round rewrites working code on
it.** The confirmation is the parent's, at acceptance: where the finding's premise is what something
outside this change does — a tool's flag semantics, a library's behaviour, a published format — the
parent checks that premise against the thing's own published documentation or source, never a
secondary summary, and records the confirmation with the round's pass log. A premise the primary
source does not support withdraws the finding with that reason through the round's auto-decisions
below, and the working code stands.

**After a fix round, re-run on deltas — every slot that raised a finding the round touched,
Critical, Important or Minor alike.** A slot's last-reviewed sha is
held **per slot per worktree**: each dispatch sets that slot's sha in every worktree to the HEAD it
was dispatched against, and a slot not dispatched in a round keeps the shas it had. A delta is
`<abs-worktree>/.superpowers/sdd/slot-delta-<round>-<slot>.diff` (the canonical worktree's), written
with `write-panel-diff.sh slot-delta <round> <slot> <abs-worktree> <worktree> <merge-base> <held-sha|-> […]`
— `-` where the slot holds no sha in that worktree, which then contributes its whole
`git diff <merge-base>` section — whose exits read as the pass-1 write's. Every slot's dispatch prompt names the path it was given and, for a delta,
each worktree's starting sha. **Check base movement first** above clears every slot's held sha in
the rebased worktree on a clean rebase, whether taken at panel entry or at a round boundary, so
that worktree's section falls under the no-held-sha rule in the next round. Then:

- **a slot re-runs when it raised a finding the round touched — Critical, Important or Minor
  alike — or the previous round raised a new Critical**; the new-Critical clause sends every slot
  in the resolved roster, Primary included, and every operator-added slot already dispatched in
  an earlier pass of this run. **A fix round is never exempt from being itself reviewed**: the
  fix for one finding can introduce the very class it fixed, no per-file or per-test check
  catches that, and the round's own mutation-proof and reproducer flips never substitute for the
  re-run its fixes name. A slot that raised nothing keeps the result it has; the round's own
  mutation-proof (below) covers what the fix changed;
- **a diff-reading slot that re-runs reads its delta**; Mutation reads no diff
  file and re-runs in its pass-1 shape, throwaway worktree included. **A diff-reading slot whose
  delta is empty in every worktree is not dispatched** — record `not re-run —
  nothing new since its last read` with `flow record pass -round <round>`;
- **a slot the operator has not named for this run is never added here** — that addition happens
  only through the explicit-request check **The roster** states, at the start of any round;
- **on a decided panel, each re-running role runs alone, in its own dispatch, on
  the decision's `panel.rerun_dispatch` pair**, under the 5-minute ceiling — never bundled with
  another role. **The re-run is targeted at what that role raised and nothing else**: in place of
  its held-sha delta it reads the round's `fix-round-N.diff` plus the sites of its own open
  findings, each opened at its recorded `file:line` in the current tree, and its prompt lists
  those `F<n>` rows verbatim, states that it is re-reviewing their fix, and names that diff path.
  Its verdict is per listed finding — fixed, or not fixed with the reproducer output — plus any
  defect the fix diff itself introduces at those sites.
  A role that raised nothing in the previous round does not re-run under the new-Critical clause
  above either — with no finding of its own to target it has nothing to re-review. Each
  `-model`/`-effort` recorded is the rerun pair (**Bundled dispatch**).

**A fix-round re-run reviews the fix, never the branch** — and no pass after pass 1 re-reads the
whole branch; the paragraph below is the one statement of a re-run's read scope. Every re-running slot's dispatch prompt —
on every panel, Mutation included — carries the FIX-ROUND SCOPE paragraph,
with `<fix report>` the round's `panel-fix-report-<round>.md` (every chunk's file on a chunked
round):

> **FIX-ROUND SCOPE:** this is a fix-round re-review. Its scope is the fix diff you were given
> and the sites of the findings it fixes, plus at most the code neighbouring those hunks — the
> enclosing function or section — and nothing else on the branch. Run only the tests and
> specs that diff touches, never the module, repository or live-spec suite. The fix report at
> `<fix report>` carries each fixed finding's proof: the test run before the fix and after it,
> and the failure with the fix reverted. Check that proof against the diff; reproduce it
> yourself only where it is missing, does not match the diff, or does not show the failure it
> claims — and say which of those it was.

**From a change's third fix round on, a fix round is scoped.** A re-running diff-reading slot
reads the round's `fix-round-N.diff` plus the sites of every finding an earlier round raised —
each site opened at its recorded `file:line` in the current tree — in place of its held-sha
delta; the re-run rule above is unchanged, and so is everything the round's own mutation-proof
covers.

**The cap check on a re-run** is `check-panel-diff-size.sh <worktree> <sha> <cap>` once per
worktree per **distinct** held sha among the diff-reading slots dispatched this round (two slots
sharing a sha in a worktree need one call there, not two); a slot with no held sha in a worktree
counts from that worktree's merge base. **The gating count is the largest per-slot sum across
worktrees** — the largest single combined read any one slot this round faces — and an exit-1
result from a call contributing to it proceeds unasked exactly as the pass-1 check does (**The
roster**, above), reporting the gating sum, its per-worktree counts and, when it differs, the
full-branch sum. Record both with `flow record pass -round <round>` for this round,
alongside the agents-ran/why/diff-path lines the fix pass records.

**The docs-only guard runs again beside that cap check**, `check-panel-docs-only.sh <worktree>
<merge-base>`. A branch that stays docs-only keeps the
reduced roster, and `primary` re-runs on its delta as above. A branch the fix round made no
longer docs-only — exit 1 or 2 where pass 1 saw exit 0 — dispatches, in this round, every
resolved slot not yet dispatched this run, each reading the whole `final-review.diff` under the
no-last-reviewed-sha rule above; Mutation among them takes its pass-1 shape, throwaway
worktree included. Record which slots joined this way and the path the guard printed.

Union all **open** findings, dedupe by **defect identity — file:line + theme.** *File:line* is the
finding's own recorded location, taken verbatim from the findings table. *Theme* is the finding's
one-sentence Note column, reduced to its own defect noun phrase — the shortest phrase naming what is
wrong, severity words and slot names stripped out.

**Before dispatching the fix subagent**, the parent runs:

```bash
check-panel-reproducers.sh <worktree> <change>
```

Exit 0 proceeds. Exit 1 covers two classes: a **missing or malformed field** is added before
dispatch; a **rejected reproducer shape** (a shell metacharacter, an absolute path, a `..` segment,
a leading `-`, a URL, a NUL byte) is a **refusal** — the line is recorded **unverifiable** and put to
the operator, never silently rewritten. Exit 2 stops the run. Both guards first read the change's
state record through the worktree argument, and a store-reached-but-absent record is exit 2 — on a
cross-repo change the record resolves only through the canonical worktree's project key, so that
shape is what invoking a guard on a peer tree reads as, and the empty findings array a peer tree's
store read would otherwise answer is never a clean verdict (KAN-658).

**The exit-code contract is checked mechanically before any dispatch decision reads a reproducer by
hand**:

```bash
check-panel-reproducer-exit-contract.sh <worktree> <change>
```

The guard runs every **open** finding's runnable reproducer through `run-reproducer.sh` against the
finding's own worktree — per finding, the worktree whose copy of the reproducer's recorded relative
path resolves: the canonical worktree first, else the one other entry of the state record's
`worktrees` map carrying the path; a path resolving in several recorded worktrees is cannot-answer,
its tree genuinely unchoosable (KAN-658) — bare, and requires the verdict *defect demonstrated* —
the exit-code behaviour an open
finding's reproducer claims on the tree under review. Before anything runs, the guard audits the
instrument itself: each runnable reproducer carries the `# demonstrates:
<path>:<line>:<content>` declaration its authoring rule above requires, within the script's first
10 lines, and the guard resolves every citation against the finding's resolved worktree — the path
stays inside the
tree, the file exists, the line exists, the content appears on that line. A reproducer whose
citation does not resolve, or whose script cannot be read to audit, is never run — a verdict spent
on an unresolvable instrument is the green flip the audit exists to deny. The audit also resolves
every `# premise:` declaration the same way, tolerantly: absence of premise lines violates
nothing, and a declared-but-unresolvable premise joins exit 1's violation classes;
mutation-declared reproducers skip it as they skip the demonstrates audit (KAN-839). Findings at any other
status claim nothing about the current tree and are skipped, as are the exemption and bare-`none`
forms the lexical guard above owns, and a mutation-declared reproducer skips the audit: what it
demonstrates is the mutated tree it builds at run time, and the `--reproducer-sha` pin below is
its instrument audit. Exit 0 proceeds to the per-finding runs below. Exit 1 names every finding
whose reproducer contradicted its claim, one disposition per class: a reproducer that read *not
demonstrated* — the inverted class —, one whose demonstrates citation did not resolve, and one
whose script cannot be read to audit at all, are
bounced exactly as the per-finding run's own answer 1 below,
once, back to the raising slot; a reproducer the runner refused as unusable is recorded
**unverifiable** and put to the operator, exactly as the per-finding run's own answer 2 below, since
a refused reproducer never ran and so carries no passing output a bounce could carry. Exit 2 — any
reproducer the runner could not verdict: a timeout, a surviving process, a plumbing failure, the
state record absent or unreadable, an ambiguous worktree — stops
the run, the same as the lexical guard's exit 2.

**For each open finding whose record carries a runnable `finding-reproducer:` command**, the
parent runs it — every finding's run in one Bash call (each throwaway copy is already
gone, removed as its slot's dispatch closed), each run followed by `; echo "F<n>: exit $?"` so every exit code stays
readable:

```bash
run-reproducer.sh <worktree> "<the finding's finding-reproducer: text>"
```

`<worktree>` is the cwd contract's per-finding resolution — the worktree whose copy of the
recorded relative path resolves, canonical first, else the one recorded worktree carrying it
(KAN-658) — the same tree the exit-contract guard above already ran this reproducer in.

Read its exit code: **0** dispatches the finding; **1** bounces it once, back to the raising slot,
carrying the reproducer's passing output; **2** is a refusal — recorded **unverifiable**, put to the
operator; **3** is a timeout or a detached survivor — recorded **unverifiable**, put to the
operator, with a surviving pid named when the script names one; **4** stops this finding's dispatch
decision entirely — a finding recorded `none — <reason>` is dispatched without a run.

**Once the fix subagent reports, re-run every dispatched finding's reproducer** under the same
constraints, carrying `--pre-fix-verdict <that finding's dispatch-time verdict>` — `demonstrated`
or `not-demonstrated`, the verdict that finding's dispatch-time run printed (its recorded exit
code names it: 0 means `demonstrated`, 1 `not-demonstrated`) — **and `--reproducer-sha <the sha
that run printed>`**: the verdict comparison is valid only between two runs of the same file, so the
runner pins the re-run to the dispatch-time reproducer and refuses a mismatch (exit 2, never
executed). A refused re-run means the reproducer was re-authored — its mutation-convention
declaration included — and it is re-run against the defect-present code — the pre-fix leg
`prove-reproducer.sh` runs, against a scratch worktree at the pre-fix commit — for a fresh
verdict and sha before the round continues: a run that answered `2`, `3` or `4` refused, went
unverifiable,
or stopped before any dispatch, so nothing re-runs for it — and
require the reproducer now to exit **0**, which the script answers **1**. The flag makes the
script refuse (exit 2) a reproducer whose verdict here is identical to its pre-fix verdict —
ambiguous under either exit-code convention, the expected one named in the script's message — so
an ambiguous reproducer is recorded unverifiable and put to the operator rather than read as an
unfinished fix. **The parent runs these re-runs itself, in its own
Bash calls** (**Dispatch sites — the
parent's closed list**, `skills/flow/implement.md`). **The flip alone does not close a finding — the fix's
diff must also touch at least one path the finding named, with a non-comment, non-whitespace
change.** A fix that meets neither that match nor one of the three shapes below is not a fix: the
finding stays open and goes through the handback below. **Beside that match, three shapes close on their own evidence, each judged from the same
fix diff or repaired history the match reads, never from the fix subagent's report alone:**
(a) **a test-only fix** — every non-context change in the diff is in a test file, a file the
project's `## test` commands run or that only test code imports — closes on the
flip the re-run already required, when the finding's pinned reproducer flipped demonstrated → not
demonstrated: the defect was a missing or weakened test, and the diff that adds or strengthens it
is the fix; (b) **a comment-only diff closes a comment-only finding** — the finding's defect is a
stale or wrong comment and the diff's change to the named path is comment-only, a repair that can
never carry the non-comment change the match demands; (c) **a history repair closes a finding
located in a commit record** — the defect lives in a commit rather than in a file, and the fix
amends, fixup-folds or rewrites the named commit under the rewrite rules above, closing on the
commit as it now stands, read post-repair: no diff-path evidence exists for a defect in history.

A re-run whose script fails a premise assertion — the loud non-zero the authoring rule requires —
reads *demonstrated*, identical to the dispatch-time verdict, and is refused as ambiguous by
design (KAN-839): the renamed premise voids the reproducer, and the refusal routes it to
re-authoring. That refusal is the fix working, never a runner bug.

A finding meeting both conditions is recorded closed there and then — a Minor recorded
`none — <reason>` has no reproducer to flip and closes on the path condition alone, or on its
comment-only alternative when the finding's defect is a comment and the diff's comment-only
repair touches that named path:

```bash
flow record status -change <name> -ref F<n> -status fixed
```

**The parent records it, never the fix subagent.** **Record every verdict this turn reached in one call,
never deferred to the round's end** — the reproducer re-runs and the fix diff are read in one
call, each finding judged, then every `status fixed` recorded together, so an aborted round
still leaves every already-verified finding closed. **A finding is recorded `fixed` only after
the re-run that verifies it — never in the same Bash call as that re-run.** The call that runs
the re-runs prints each exit and is read; a separate later call records `status fixed` for
exactly the findings it verified, and no others. **A finding failing
the match and all three shapes is left untouched** on `open`, for the handback below. This walk follows
**Read discipline** (`skills/flow/implement.md`): the specific hunks a finding names, never the <!-- refs-guard:allow -->
whole fix diff.

### The fix round mutation-proves what it changed

**Every executable behaviour the fix changed is mutation-proved, not only the test cases the round
adds.** The fix subagent performs the proof and reports it, per the MUTATION PROOF paragraph its
dispatch carries; the parent runs no build of its own here. A survivor the fix subagent cannot
judge real or equivalent goes through the same handback the section already names.

**A fix that adds or strengthens a test is proved by a flip of the fixed line itself, one flip per
fixed finding.** The mutation flips the line the fix changed — the code the added or strengthened
test exists to catch — and the test confirmed to fail under the flip is that added or strengthened
test, not merely any test that happens to trip: the flip is what shows the fix's own test tests
what it claims. The flip is recorded with the round's other `flow record mutation` rows before the
round closes.

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

**A fix to a guard script is proved by the guard's own observable, measured on both sides of the
fix.** A guard script is a check that exists to fail on a defect class — this project's
`scripts/check-*` family, or wherever its `## lint` section names its guards. The mutation probe
is the guard run itself against the defect state: on the pre-fix code the guard reports the
defect — the probe fails — and on the post-fix code it does not — the probe passes. The
`fix-mutation:` line for that behaviour carries the measured pre/post observable as its third
field, `<pre>→<post> <what the observable counts>`
(`2→0 orphaned temp lists`), never a bare test name.

**The parent checks the reported list against the fix diff before the round can close, reading the
`fix-mutation:` lines and walking the diff itself.** Walk every
hunk of the fix diff with a non-comment, non-whitespace change: each one is either covered by a
reported line, or is not an executable behaviour at all. A fixed finding whose fix added or
strengthened a test closes only with its own flip in the pass log — the per-finding obligation
above is checked beside this walk, and a finding short its flip stays open for the handback below,
never closed on another finding's line. A hunk that removes or weakens a test or an
assertion states in the record what it used to cover and names what still covers that same behaviour
now — checked by running the named covering test against the **pre-fix** code and confirming it
fails. A hunk whose path is draw/geometry code is held to the **PIXEL PROBE** paragraph
beside the mutation lines: the fix's report names the probe assertion it landed against the
actual rendered pixels or geometry, and a fix commit for this class of bug that landed no such
probe is rejected at the fix step — the round does not close and the finding stays open, for the
handback below — rather than the regression being discovered next round. **The same
walk holds the fix subagent to its PLAN FIELDS obligation:** a hunk that adds a
test case, adds a file, or changes what a task's `**Baseline:**` counts, whose task's `**Tests:**`,
`**Baseline:**` or `**Files:**` field in `<changeRoot>/tasks.md` does not reflect it, does not close
the round; it goes to the handback.

Beside the reproducer re-runs above, the round close re-runs the task-field guard mechanically:
`check-task-commit-fields.sh <worktree> <task-id> <task-sha> "" <canonical-worktree>
<name>` for every task a fixup folded into — `<task-id>` from that commit's `Task-Id:` trailer,
`<task-sha>` the folded commit as it now stands, the remaining arguments resolved the way
`skills/flow/implement.md`'s task-close step resolves them. Exit 1 does not close the round; it
goes to the handback. Exit 2 — the guard's not-a-verdict close (the same course
`skills/flow/implement.md`'s task-close step gives it) — stops the run: a handback cannot repair
an inability. The stale-field classes are the guard's verdicts: a task declaring `**Baseline:** before=N
after=M` fails when the changed files' `@Test` delta at the commit does not measure it — skipped
where the counted set carries no `@Test` at either revision, per the skip-not-fail rule — and a
`**Tests:**` name, backticked or bare camelCase, that the tree's content at the commit no longer
contains fails with it. A test added to the commit with no `**Baseline:**` declared and no
`**Tests:**` naming it stays the walk's judgment.

**The round close runs the project's configured build-green guard too, when the project declares
one (**The guard's scope**, `skills/flow-contracts/build-green.md`), over the plan's `tasks.md` —
exit 1 does not close the round; it goes to the handback, and exit 2 — the same not-a-verdict
close — stops the run.** This is the gate that holds a
plan appended to mid-run to the tags it published under: tasks a fix round appends carrying
`**Build:** pending` keep the round open until every tag reads `green` or `red`.

This binds the fix round every run — the obligation is the round's, not a slot's.

**Rebuild the dispatch context bundle before dispatching the fix subagent**, same as above,
overwriting the same path. Report the script's stderr line (`bundle unchanged — reusing …` or
`bundle rebuilt — …`) as part of this round's own reporting.

**Carry each surviving finding to the fix subagent as a structured block**, not a bare restatement
of its prose: its `F<n>`, the slot that raised it, its severity, its `file:line`, its theme, the
text of its `finding-reproducer:` line, its slot's report path, and any bounce already recorded
against its defect identity. **Inline no source excerpt.**

### The fix dispatch

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
`ceil(findings raised in earlier rounds / 10)`. Before recording each `dispatch begin`, confirm that key has
no panel-fix begin already recorded — a second begin under a fresh key is over-dispatching even
when every key is well-formed. `check-panel-fix-single-dispatch.sh` holds every run's panel close
to exactly this shape (**Before closing the stage**, below). Inline
(`skills/flow/implement.md`'s **Inline — the parent implements**), the parent applies the fix
itself under the same paragraphs, dispatching no subagent, and records the pass with `-role
panel-fix -model <parent model> -effort <parent effort> -agent-id inline`, `begin` before its first
edit and `end` once the fix commit lands, as that section's **Records** bullet states. Where a finding is
confirmed as a real defect, the fix subagent invokes **superpowers:systematic-debugging** before
writing its fix. **Dispatch it on the decision's `fixer` object** — its own model and effort,
chosen apart from the implementer's, `subagent_type:
flow-<effort>` with that `model` passed as the Agent tool's own `model` parameter, and
`-model`/`-effort` below carry that pair. On harness `zcode` the pair given and recorded is `glm-5.3-flash` / `high` instead (**Harness mapping**, `skills/flow-contracts/model-policy.md`). Record
every pass with `flow record pass -round <round>`: which agents ran, why,
the diff path they read, and — when this pass bounced any finding — each bounced finding's defect
identity together with the reproducer output it carried back.

A Minor fixed or withdrawn blocks nothing; a Minor left `open` or `deferred` blocks exactly as a
Critical does. When fix rounds do not converge, the run decides each unresolved finding itself, one finding
at a time, and does not ask:

- **A defect a code or document change can resolve takes another fix round** — at Critical or
  Important, and even where the fix reaches past the change's original scope. A Minor that reached
  this loop joins that round only beside a Critical or Important taking one; with none, the parent
  fixes it inline, as **Panel re-runs** (`skills/flow/review-panel.md`) fixes a Minor-only round's
  Minors. No round count ends this: a round that made no progress on a defect takes **Fewest
  operator actions** (`skills/flow-contracts/pipeline.md`).
- **A finding no change to the tree can resolve is withdrawn** — a verification-only ask, a proof
  that needs an environment the run does not have, a defect something already covers — recorded
  `-status 'withdrawn <reason>'`, the reason one clause naming that mechanism. That reason stands
  where the operator's would.
- **Only a genuine inability reaches the operator:** a finding the run cannot judge either way, or
  a fix that needs an irreversible or outward-facing action. Only then is the prompt below raised, shape per Operator prompts
  (`skills/flow-contracts/operator-prompts.md`), and resolved per that contract's **Auto-resolution**
  and **What still stops**: its recommended option is taken unasked, and asked only where the fix
  is irreversible or outward-facing:

> **`<location>` — <the finding, in one line>. The fix round did not resolve it.**
> - **Take another round on it** *(default, recommended)*
> - **Withdraw it — I'll give the reason** — the reason is recorded on the finding's marker line
> - **Stop the run and hand it back to me**

**Every automatic decision is recorded in the pass log** with `flow record pass -round <round>
-note 'auto-decided F<n>: another round — <the defect>'` or `-note 'auto-decided F<n>: withdrawn —
<reason>'`, so the handoff shows it. Only this loop records `withdrawn`: automatically with the
run's reason, or on the operator's answer with theirs.

**A fix round closes on a clean re-run, never on its own verification.** The reproducer re-runs
and the fix-diff path check above close the findings that fix addressed; they never close the
panel. After a round that dispatched a `panel-fix` chunk, the delta re-run the rules above trigger
runs before the close guards below do, and the stage may close only on a re-run that raised no new
finding above Minor; an empty-delta re-run whose slots were recorded `not re-run — nothing new
since its last read` counts as clean, and a re-run that re-raises a defect withdrawn
under the handback above is recorded `withdrawn` with its original reason, never
`open`, and does not stand in the way of that clean round. The last round before the stage close
is therefore one of: a pass 1 that raised nothing, a re-run that raised nothing, a re-run whose
only raise was a defect recorded `withdrawn` under the carve-out above, or a round whose findings
were all Minor and none fixed — nothing it touched names a re-run — which closes beside them. A
re-run that finds a fix incomplete opens the next fix round under the rules above, and the cycle repeats until a
re-run comes back clean.
