# Review panel — fix rounds

Loaded during `flow.review` by the load directive of **Panel re-runs**
(`skills/flow/review-panel.md`), only once a round has recorded a Critical or Important finding, or read by section where a citer outside the panel stage names it.
Every section name below without a path is a section of this file or of
`skills/flow/review-panel.md`; the fix subagent's dispatch paragraphs and its dispatch record stay
there.

## Panel re-runs

**Pass 1 runs the decided panel plus every slot the operator named at this stage's start that it did not
already carry; a fix's review is one reviewer (`skills/flow/fix.md`, step 4).** Only re-runs after a fix are scoped. Record
`FIX_BASE` — the branch tip the fix round starts from, per worktree — commit the fix, then write
`<abs-worktree>/.superpowers/sdd/fix-round-N.diff` and its touched list with
`flow-guard write-panel-diff fix-round <N> <abs-worktree> <worktree> <fix-base> [<worktree> <fix-base>…]`,
whose exits read as the pass-1 write's (`skills/flow/review-panel.md`).

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
  catches that, and nothing substitutes for the re-run its fixes name — for every role but
  Mutation, below. A slot that raised nothing keeps the result it has;
  **Mutation is the one role never re-dispatched after a fix.** A surviving mutant the round fixed
  closes when the fix adds the test the finding named. A round whose every Critical and
  Important is a surviving mutant therefore dispatches its fix and closes with no re-run round;
  Mutation sweeps again only at the next pass 1.
- **a diff-reading slot whose
  delta is empty in every worktree is not dispatched**;
- **a slot the operator has not named for this run is never added here** — that addition happens
  only through the explicit-request check **The panel** states, at the start of any round;
- **the round's re-running roles ride one bundled dispatch, on role `panel-rerun`**, its pair
  from `flow-guard pair-judge` — one **PASS** section
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

### What a fix round leaves behind

**Every fixed Critical or Important finding leaves a tracked suite test behind** — the
`regression-test:` line of the REPORT FILE paragraph (`skills/flow/review-panel.md`), so a
regression of a reviewed defect is still caught once the worktree is gone; a fixed Minor needs none.
The parent checks the fix diff against the report with `git diff --stat` and the hunks a finding
names. A fix to draw/geometry code is held to the PIXEL PROBE paragraph of
`skills/flow/fixer-brief.md`: a fix commit that landed no probe against the rendered pixels is
rejected at the fix step — the round does not close and the finding stays open, for the handback
below.

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
(`panel-fix-<round>-2`, `-3`, …), contiguous from 2. A gated reviewer's fix round is no `panel-fix-<round>` key: it records under its own key, `-role panel-fix` (`skills/flow/gated-review-fix.md`), and a fix's implementer under `fix-<k>-implementer` (`skills/flow/fix.md`). Where a finding is
confirmed as a real defect, the fix subagent reads and follows **superpowers:systematic-debugging**
from its file before writing its fix — on Claude Code the newest
`~/.claude/plugins/cache/*/superpowers/*/skills/systematic-debugging/SKILL.md`, since a `flow-*` agent
carries no Skill tool. **Dispatch it on the `panel-fix` role's pair** from `flow-guard pair-judge` (**Model and effort**,
`skills/flow/brainstorm-planner.md`), `subagent_type` the pair's agent type with its `model` passed as
the Agent tool's own `model` parameter, and `-model`/`-effort` below carry that pair.

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
