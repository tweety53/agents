# Review panel — fix rounds

Loaded during `flow.review-panel` by the load directive of **Panel re-runs**
(`skills/flow/review-panel.md`), only once a round has recorded a Critical or Important finding, or read by section where a citer outside the panel stage names it.
Every section name below without a path is a section of this file or of
`skills/flow/review-panel.md`; the fix subagent's dispatch paragraphs and its dispatch record stay
there.

## Panel re-runs

**Pass 1 runs the panel **The docs-only reduction** chose — the decided panel, or `primary` alone
on a docs-only branch — plus every slot the operator named at this stage's start that it did not
already carry; a fix run's pass 1 runs the panel **The fix-run review** chose.** Only re-runs after a fix are scoped. Record
`FIX_BASE` — the branch tip the fix round starts from, per worktree — commit the fix, then write
`<abs-worktree>/.superpowers/sdd/fix-round-N.diff` and its touched list with
`flow-guard write-panel-diff fix-round <N> <abs-worktree> <worktree> <fix-base> [<worktree> <fix-base>…]`,
whose exits read as the pass-1 write's (**The docs-only reduction**, `skills/flow/review-panel.md`).

**A branch the remote already holds takes the fix as one new commit on top, never a rewrite.**
This is the normal case: the branch is pushed with every commit (**Branch backup**,
`skills/flow-contracts/git-boundaries.md`), so the fix stages the changed paths
(`git add -- <the changed paths>` — a pathspec commit reads tracked paths only) and is a plain
`git commit -m ... -- <the changed paths>` at the tip — the pathspec-scoped default (**Branch backup**, `skills/flow-contracts/git-boundaries.md`),
so it carries only the paths the finding named, whatever else the index holds — pushed
plain like any other commit, and every downstream commit keeps its sha. The commit carries the round's `Flow-Key:` trailer — `git commit -m <subject> -m "Flow-Key: <the round's dispatch key>" -- <the changed paths>` (**Commits outside every task**, `skills/flow-contracts/git-boundaries.md`). Before it is pushed, the untasked-commit check of **Commits outside every task** (`skills/flow-contracts/git-boundaries.md`) runs.

**Which slots re-run, and on what, follows from the findings the round fixed and the severities
it raised.**

**A Critical is confirmed against a primary source before any fix round rewrites working code on
it.** The confirmation is the parent's, at acceptance: where the finding's premise is what something
outside this change does — a tool's flag semantics, a library's behaviour, a published format — the
parent checks that premise against the thing's own published documentation or source, never a
secondary summary, and records the confirmation with the round's pass log. A premise the primary
source does not support withdraws the finding with that reason through the round's auto-decisions
below, and the working code stands.

**After a fix round, re-run on deltas — only the slots that raised a finding the round touched,
Critical, Important or Minor alike.** Then:

- **a slot re-runs when it raised a finding the round touched — Critical, Important or Minor
  alike**, a row linked by `-supersedes` to another
  role's row of the same round counting as that other role's (**Bundled dispatch**, `skills/flow/review-panel.md`). **A fix round is never exempt from being itself reviewed**: the
  fix for one finding can introduce the very class it fixed, no per-file or per-test check
  catches that, and the round's own mutation-proof never substitutes for the
  re-run its fixes name — for every role but Mutation, below. A slot that raised nothing keeps the
  result it has; the round's own mutation-proof (below) covers what the fix changed;
  **Mutation is the one role never re-dispatched after a fix.** A surviving mutant the round fixed
  closes when the fix's `fix-mutation:` lines prove the tests it added. A round whose every Critical and
  Important is a surviving mutant therefore dispatches its fix and closes with no re-run round;
  Mutation sweeps again only at the next pass 1.
- **a diff-reading slot whose
  delta is empty in every worktree is not dispatched** — record `not re-run —
  nothing new since its last read` with `flow record pass -round <round>`;
- **a slot the operator has not named for this run is never added here** — that addition happens
  only through the explicit-request check **The panel** states, at the start of any round;
- **the round's re-running roles ride one bundled dispatch, on role `panel-rerun`**, its pair
  from `flow-guard pair-judge`, under the 5-minute ceiling — one **PASS** section
  per role, the one-per-role fallback (**Bundled dispatch**, `skills/flow/review-panel.md`)
  separating them one per role when the bundled findings need separation. **Each role's re-run is
  targeted at what that role raised and nothing else**: it reads the round's `fix-round-N.diff` plus the sites of its own open
  findings, each opened at its recorded `file:line` in the current tree, and its PASS section lists
  those `F<n>` rows verbatim, states that it is re-reviewing their fix, and names that diff path.
  Its verdict is per listed finding — fixed, or not fixed — plus any
  defect the fix diff itself introduces at those sites.
  Each
  `-model`/`-effort` recorded is the rerun pair (**Bundled dispatch**).

**A fix-round re-run reviews the fix, never the branch** — and no pass after pass 1 re-reads the
whole branch; the paragraph below is the one statement of a re-run's read scope. Every re-running slot's dispatch prompt,
on every panel, carries the FIX-ROUND SCOPE paragraph,
with `<fix report>` the round's `panel-fix-report-<round>.md` (every chunk's file on a chunked
round):

> **FIX-ROUND SCOPE:** this is a fix-round re-review. Its scope is the fix diff you were given
> and the sites of the findings it fixes, plus at most the code neighbouring those hunks — the
> enclosing function or section — and nothing else on the branch. Run only the tests and
> specs that diff touches, never the module, repository or live-spec suite. The fix report at
> `<fix report>` carries each fixed finding's proof: the test run before the fix and after it,
> and the failure with the fix reverted. Check that proof against the diff; reproduce it
> yourself only where it is missing, does not match the diff, or does not show the failure it
> claims — and say which of those it was. Read every hunk of the fix diff, a fix made for another
> slot's finding included, for a defect the fix itself introduces — an evaluation moved out of the
> guard that kept it safe, a branch the fix changed — and report it as a new finding at its line.
> Given a diff, your report file carries one verdict line
> per finding you were sent to re-review, in exactly this shape: `verdict: F<n> — fixed —
> <path>:<line>` or `verdict: F<n> — not fixed — <path>:<line>`, where `<path>:<line>` is a line
> inside the hunk of that diff you read for that finding (`none` in its place, on a `not fixed`
> line, when the diff touched nothing of it).

**The docs-only guard runs again**, `flow-guard check-panel-docs-only <worktree>
<merge-base>`. A branch that stays docs-only keeps the
reduced panel, and `primary` re-runs on its delta as above. A branch the fix round made no
longer docs-only — exit 1 or 2 where pass 1 saw exit 0 — dispatches, in this round, every
decided slot not yet dispatched this run, each reading the whole `final-review.diff`; Mutation among them takes its pass-1 shape, throwaway
worktree included. Record which slots joined this way and the path the guard printed.

Union all **open** findings, dedupe by **defect identity — file:line + theme.** *File:line* is the
finding's own recorded location, taken verbatim from the findings table. *Theme* is the finding's
one-sentence Note column, reduced to its own defect noun phrase — the shortest phrase naming what is
wrong, severity words and slot names stripped out.

**The fix's
diff must touch at least one path the finding named, with a non-comment, non-whitespace
change.** A fix that meets neither that match nor one of the three shapes below is not a fix: the
finding stays open and goes through the handback below. **Beside that match, three shapes close on their own evidence, each judged from the same
fix diff or repaired history the match reads, never from the fix subagent's report alone:**
(a) **a test-only fix** — every non-context change in the diff is in a test file, a file the
project's `## test` commands run or that only test code imports — closes: the defect was a missing or weakened test, and the diff that adds or strengthens it
is the fix; (b) **a comment-only diff closes a comment-only finding** — the finding's defect is a
stale or wrong comment and the diff's change to the named path is comment-only, a repair that can
never carry the non-comment change the match demands; (c) **a history repair closes a finding
located in a commit record** — the defect lives in a commit rather than in a file, and the fix
amends, fixup-folds or rewrites the named commit, closing on the
commit as it now stands, read post-repair: no diff-path evidence exists for a defect in history.

A finding meeting that match or one of the three shapes is recorded closed there and then:

```bash
flow record status -change <name> -ref F<n> -status fixed
```

**The parent records it, never the fix subagent.** **Record every verdict this turn reached in one call,
never deferred to the round's end** — the fix diff is read in one
call, each finding judged, then every `status fixed` recorded together, so an aborted round
still leaves every already-verified finding closed. **A finding failing
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
round's re-review begins.

**Every fixed Critical or Important finding leaves a tracked suite test behind** — the
`regression-test:` line of the REPORT FILE paragraph (`skills/flow/review-panel.md`), so a
regression of a reviewed defect is still caught once the worktree is gone; a fixed Minor needs none.

**A fix round that fixed a Critical or Important finding and changed no executable behaviour still
records one `-mutated none -test <reason>` row.**

**Record each one with one `flow record mutation -change <name> -round <round> -path <path>
-mutated <what> -test <test>` call per line, transcribed from the fix subagent's final message — the
exemption form records `-mutated none -test <reason>`; the lines the calls write are:**

```text
fix-mutation: <path> — <what was mutated> — <the test that failed>
fix-mutation: <path> — none — <reason>
fix-mutations-total: <n>
```

Each row is recorded before the round's re-review begins.

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
actual rendered pixels, and a fix commit for this class of bug that landed no such
probe is rejected at the fix step — the round does not close and the finding stays open, for the
handback below — rather than the regression being discovered next round.

This binds the fix round every run — the obligation is the round's, not a slot's.

**Carry each surviving finding to the fix subagent as a structured block**, not a bare restatement
of its prose: its `F<n>`, the slot that raised it, its severity, its `file:line`, its theme, and
its slot's report path. **Inline no source excerpt.**

### The fix dispatch

Give the surviving findings to fix subagents in **chunks of at most 10 findings**. The round's
findings are split into sequential chunks of at most 10 — the cap that keeps one dispatch from
re-reading the whole branch across every finding — each chunk one panel-fix dispatch carrying its chunk of
the combined list. Never one dispatch per reviewer, per slot, or per finding — that split
fragments one diff into competing fixups against the same worktree. Chunks run in order, each
launched with `run_in_background: true` and its completion awaited at the turn's end (Turn
discipline, **4. Execute (SDD + TDD)**, `skills/flow/implement.md`) before the next chunk begins. `<round>` is the fix
round, one above the review round whose findings it fixes: the fix after `panel-0` is
`panel-fix-1`. A fix commit's
`Flow-Key:` trailer names that same fix round. The round's first
chunk's `-key` is exactly `panel-fix-<round>`; each further chunk appends `-<n>`
(`panel-fix-<round>-2`, `-3`, …), contiguous from 2. A gated reviewer's fix round is no `panel-fix-<round>` key: it records under its own key, `-role panel-fix` (`skills/flow/gated-review-fix.md`), and an in-run verify fix under its appended task's implementer pair (`skills/flow/verify-fix-loop.md`). Where a finding is
confirmed as a real defect, the fix subagent reads and follows **superpowers:systematic-debugging**
from its file before writing its fix — on Claude Code the newest
`~/.claude/plugins/cache/*/superpowers/*/skills/systematic-debugging/SKILL.md`, since a `flow-*` agent
carries no Skill tool. **Dispatch it on the `panel-fix` role's pair** from `flow-guard pair-judge` (**Model and effort**,
`skills/flow/brainstorm-planner.md`), `subagent_type` the pair's agent type with its `model` passed as
the Agent tool's own `model` parameter, and `-model`/`-effort` below carry that pair. Record
every pass with `flow record pass -round <round>`: which agents ran, why,
and the diff path they read.

A Minor fixed or withdrawn blocks nothing; a Minor left `open` or `deferred` blocks exactly as a
Critical does. When fix rounds do not converge, the run decides each unresolved finding itself, one finding
at a time, and does not ask:

- **A defect a code or document change can resolve takes another fix round** — at Critical or
  Important, and even where the fix reaches past the change's original scope. A Minor that reached
  this loop joins that round only beside a Critical or Important taking one; with none, the
  `minor-fix` subagent fixes it, as **Panel re-runs** (`skills/flow/review-panel.md`) fixes a
  Minor-only round's Minors. No round count ends this: a round that made no progress on a defect takes **Fewest
  operator actions** (`skills/flow-contracts/pipeline.md`), as does a fourth round after three
  failed ones (**Three failed rounds**, there).
- **A finding no change to the tree can resolve is withdrawn** — a verification-only ask, a proof
  that needs an environment the run does not have, a defect something already covers — recorded
  `-status 'withdrawn <reason>'`, the reason one clause naming that mechanism. A "something already covers" reason that claims a
  test passes names the run — its command and exit — else the finding is carried to the next
  fixer, and `flow record status` refuses a green withdrawal with no exit. That reason stands where the operator's would.
- **Only a genuine inability reaches the operator:** a finding the run cannot judge either way, or
  a fix that needs an irreversible or outward-facing action. **The round asks once.** It never
  fires an operator ask where a handback is discovered: the walk that produced it — the mutation walk — runs to its end first, every handback it
  produced is collected, and the collection is presented as one batched ask (**Batched asks**,
  `skills/flow-contracts/operator-prompts.md`) — one entry per handback, each entry carrying the
  finding's observed/breaks/fix summary in one line and its recommended disposition. Each entry
  resolves per that contract's **Auto-resolution** and **What still stops**, per entry: its
  recommended option is taken unasked, and only the entries auto-resolution leaves asked — the fix
  irreversible or outward-facing — join the ask. An entry
  reads:

> **`<location>` — <the finding, in one line: what was observed, what it breaks, the fix it needs>. The fix round did not resolve it.**
> - **Take another round on it** *(default, recommended)*
> - **Withdraw it — I'll give the reason** — the reason is recorded on the finding's marker line
> - **Stop the run and hand it back to me**

**Every automatic decision is recorded in the pass log** with `flow record pass -round <round>
-note 'auto-decided F<n>: another round — <the defect>'` or `-note 'auto-decided F<n>: withdrawn —
<reason>'`, so the handoff shows it. Only this loop records `withdrawn`: automatically with the
run's reason, or on the operator's answer with theirs.

**A fix round closes on a clean re-run, never on its own verification.** The fix-diff path check
above closes the findings that fix addressed; it never closes the
panel. After a round that dispatched a `panel-fix` chunk, the delta re-run the rules above trigger
runs, and the stage may close only on a re-run that raised no new
finding above Minor; an empty-delta re-run whose slots were recorded `not re-run — nothing new
since its last read` counts as clean, and a re-run that re-raises a defect withdrawn
under the handback above is recorded `withdrawn` with its original reason, never
`open`, and does not stand in the way of that clean round. The last round before the stage close
is therefore one of: a pass 1 that raised nothing, a re-run that raised nothing, a re-run whose
only raise was a defect recorded `withdrawn` under the carve-out above, or a round every one of
whose findings was Minor, which re-runs no slot — the `minor-fix` subagent fixes them at the
round's close (**Panel re-runs**, `skills/flow/review-panel.md`) — and closes beside them. A
re-run that finds a fix incomplete opens the next fix round under the rules above, and the cycle repeats until a
re-run comes back clean.
