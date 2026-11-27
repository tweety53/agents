# flow — rationale

Reasoning behind `skills/flow/`'s phase files: why a rejected alternative was rejected, which
design.md decision a passage implements, and choices a task made on its own. Moved here verbatim
from the run-loaded files; **no `/flow` run loads this file.** Each heading names the source file
and the section the passage came from.

## SKILL.md — Stage keys

Every stage `/flow` marks uses a `flow.*` key, minted fresh for this command rather than reusing
`start.*`/`do.*`/`finish.*` — the old namespacing was tied to the three commands this one replaces,
and reusing it here would misdescribe a stage that no longer runs under the command its key names.
**This is a design decision this task makes, not one resolved upstream of it**: the alternative —
keeping the old keys, unioned, as `/flow-fast` did — was rejected because several stages here are
not the old stage unchanged (`flow.verify` merges two, `flow.landing-routes` absorbs two more,
`flow.kickoff` is new), so a reused key would sometimes mean something different than its history
records.

## SKILL.md — preamble

> one state file, the same three states, content reorganized by topic rather than by the old
> start/do/finish boundary. This file is the router: it

## SKILL.md — Guardrails

> **Generate this run's session token once, right here, before the first mark any phase file below
> makes — a short, unique literal string — and reuse that exact same value at every `stage begin` this
> run makes, including inside every phase file it dispatches into.** One run, one token, never a fresh
> one per mark or per phase file.

## brainstorm.md — preamble

> Superpowers Basic Workflow steps **#1** (brainstorming) and **#3** (writing-plans), intertwined
> with spectre artifact creation, run here — without an options-question round or a proposal
> publish.

## brainstorm.md — A. Resolve the change and write `STARTED`

> `STARTED` is a kickoff marker,
> "the operator started this," not a record that a design was approved and a proposal published.

> `artifactUrl` stays `null` —
> `/flow` publishes no proposal artifact.

## integrate.md — preamble

> Run 1, with two folds:
> the In Review move (the Jira "move to In Review" step was a sub-step of `flow.landing-routes`; In Review has since been dropped altogether, and the self-review now runs after landing) and the `write-in-progress` fold
> (below).

## integrate.md — rebase

> — the rebase never runs on its own, only after the
> operator picks this option, and only against the worktree(s) that actually moved

Superseded: the rebase now runs with no prompt on every `MOVED` worktree — **Sync the branch onto
the base** (`skills/flow/sync-onto-base.md`).


## integrate.md — landing routes

> This stage carries three sub-steps under one mark: the git route, the state write, and
> the Jira transition — three sub-steps recorded as one.
> **The choice**: fold `write-in-progress` in alongside `move-in-review`, rather than
> leave it standalone. The write is a genuine no-op (`IN_PROGRESS` → `IN_PROGRESS`, nothing changes
> but `prUrl` and `updatedAt`/`updatedBy`) and by the
> time this mark reaches it, the route sub-step immediately above has already decided what `prUrl`
> becomes — folding the write in with the route that produces its one real input, and with the Jira
> transition that only ever follows a successful route, keeps one mark's three sub-steps in the causal
> order they already have to run in, rather than three marks whose middle one records nothing a reader
> could not already infer from the other two.

## brainstorm.md — the worktree is created inside `flow.kickoff`

An earlier design moved worktree creation to the end of planning, so sections **B**, **C** and **D** wrote into the main checkout's own
`<project>/spectre/changes/<name>/` and the run moved that directory into the worktree afterwards. That
decision is superseded by an explicit operator instruction: `/flow`, `/flow-fast` and `/flow-plan`
runs kept touching whatever branch the main checkout sat on — planning artifacts left there when a
run stopped early, a `/flow-plan` commit on the checkout's current branch — and the operator asked
that every run "always create a worktree in the very beginning and operate only on worktrees, not
touching main/develop/regular branches." Creating the worktree inside `flow.kickoff`, right after
the `STARTED` write, is what makes that hold by construction: no later phase has a main-checkout
path to write to. The interim `.superpowers-sdd-decision.json` path and the two `mv` lines that
existed only to bridge the late creation went with it. **Considered:** writing planning output to a
scratch directory and moving it in later — rejected the same way the earlier design rejected it, and now with
no gap left for it to bridge.

## brainstorm.md — Run brainstorming and planning through the planner role

The ban on every dispatch from kickoff to the plan gate narrowed to forks. Its recorded reason was about forks alone: a fork copies the whole
conversation, so it saves no context and doubles the spend. A fresh Explore or Plan agent starts
empty and returns only its conclusions, so it saves the planning session's context — the opposite
of the failure the ban was written against. **Considered:** keeping the ban for every dispatch —
rejected because fresh agents save the context the ban protected.

**Superseded (below):** the two
decisions that follow no longer hold — planning runs in a `planner` dispatch without plan mode.

On Claude Code `ExitPlanMode` is the design approval and the merged convergence confirm still
closes the checklist. **Considered:** `ExitPlanMode` in
addition to the merged confirm's approval — rejected as two approvals of one design;
`ExitPlanMode` replacing the plan review gate — rejected because that gate reviews `tasks.md` and
the decision, which do not exist yet at design time.

Planning keeps every current practice and gate and adds the plan-mode practices. **Considered:** replacing brainstorming with plan mode — rejected by the
operator.

The design approval is the convergence confirm asked alone as an **AskUserQuestion** of two
options, approve or another round, whose approve option's `preview` shows the full design text;
the session never enters plan mode. It gives the
operator plan-mode approval's read-the-whole-plan-then-approve step without the mode.
**Considered:** `ExitPlanMode` as the approval — rejected because it needs the session to be in
plan mode, a rejection strands the session there, and a subagent's plan mode is ignored under the
auto, acceptEdits and bypass permission modes, so a `planner` dispatch cannot rely on it.

## git-boundaries.md — Branch backup

Pushing the branch at creation and after every commit was asked for in the same instruction, so a
worktree lost with the machine, or removed by a stray cleanup, is rebuilt from `origin/<branch>`
rather than lost. The cost accepted: one push per commit, and integrate's push after the
reshape becomes `--force-with-lease`, on a branch only the run writes.

## cleanup.md — Worktree cleanup

Moved verbatim from the section's opening sentence, where it followed "**Worktree cleanup**
(`skills/flow-contracts/finish-contract-run2.md`), canonical for it": , and is not restated in full
here beyond one override.

## cleanup.md — Guardrails

Moved verbatim, the bullet addressed to the file's editor rather than to a run:

- **Never** state a cleanup rule here — **Temporary artifacts registry** (`skills/flow-contracts/artifacts-registry.md`)
  and **Worktree cleanup** (`skills/flow-contracts/finish-contract-run2.md`) are canonical.

## Recorded reasons

Each passage below sat inline in a run-loaded phase file beside the rule it motivated; the rule and its in-sentence reason stay there. The quoted fragment before each dash is that rule, for attribution.

### brainstorm-planner.md — The checklist

- ***A frame-specified design needs its handoff assets reachable from the tree.***  — handoffs outside the tree leave every "different from the mockup" report unanswerable.

- ***Every ambiguity is classified, and only an `[architectural]` or `[impactful: <criterion>]` one is asked.*** — asking every ambiguity, however small, filled planning rounds with questions whose recommended option the operator took every time; the classifier keeps the ones whose wrong pick is costly or that no option can honestly settle.

### brainstorm-planner.md — Decisions

- ***ID** is assigned once, at creation, and is **immutable** — the match key a later round uses to **supersede** a decision.* — The same key is what lets a review panel name the specific decision a finding found stale instead of re-arguing the design.

### brainstorm-planner.md — D. Writing plans

- ***A task cites the decision it implements, never restates it.** Restated decision prose drifts from its entry the first time either is edited* — which is why a plan cites instead
- ***End a removal task on an absence sweep.*** — so each task's done is mechanical rather than a judgment call.

### brainstorm-planner.md — Frame colours and computed rules

### brainstorm-planner.md — D. Writing plans (live check)

- ***Write a live-verification task when the change touches a running service or persistent state.** … the justification is required, the task is not.* — This is deliberately not an integration-test stage on every change: a step that usually resolves to "nothing to do" trains everyone to skip it.
- *…never a live-verification task …* — such a task repeats, against a stack of its own, what `flow.verify` and visual verify already do against the running app.
- *Write every click and capture in an end-to-end or fidelity spec on a structural locator* — A re-anchor probe after every layout-shifting task was rejected: it only detects the drift a locator removes.

### verify-and-handoff.md — Inline verify

- *The test group skips what the list's last green run on a clean tree already covered* — The skip keys on the tree, never on time: a green sha is reused only for a command whose declared inputs the delta leaves untouched.

### visual-verify.md — The verifier dispatch

- ***A final report that is a block closes its dispatch `-outcome blocked -cause <cause>`, never `completed`.*** — The cause is what makes environment-caused blocks a query the store answers instead of a footnote in one run's ledger.

### visual-verify.md — A pass is one verifier per worktree, whatever its frame count

- *…whatever its frame count* — A verifier that judges each frame from its composite and its `over` rows and reports only differences reads 30 frames in about two minutes, less than one shard of a sharded pass took over the same frames; shards added a `close` per slice, record parts to merge and per-shard keys, and every extra row is one more that can be left open.

### visual-verify.md — Visual verification, step 3 (pre-flight)

### visual-verify.md — Visual verification, step 6 (no fingerprint)

- *The stack freshness check runs once, in verify's test group* — the per-stage fingerprint and its `NEXT: restart-stack` branch were deleted; a stack visual verify starts is built from the current tree, `fix.md`'s step 3 re-runs the lint/test list when a visual fix changed code.

- *Steps 1, 2, 3, 5, 6, 12 and 13 are the parent's* — Those steps run in the parent's own calls to `flow-guard visual-verify-pass prepare` and `close` — the same reason keeps them out of the verifier — and the parent still starts, restarts and stops the stack itself, since `start` carries placeholders and `## run` is prose no script can run.

### visual-verify.md — Visual verification, step 10 (sidecar)

- ***No sidecar is never a silent skip.*** — A mockups directory can sit unused, `mockups: no map` reported and accepted every round, when nothing requires the sidecar that triggers the compose step to exist.

### visual-verify.md — Visual verification, step 11 (record)

### visual-verify.md — Visual verification, step 12 (commit)

### visual-verify.md — Visual verification (Blocking)

- ***a defect the verifier reports in a captured screenshot — even when every assertion passed.*** — That last one is the whole point of this stage: a defect can be invisible to a diff, a review panel and the test suites, and obvious the moment the page is opened.

### visual-verify.md — Visual verification, step 10 (frame-list reconciliation)

- *The frame-list half of that reconciliation is a guard* — `flow-guard visual-verify-pass close --check-only` derives the declared list from the mockups directory's frames `design.md` names, so no second copy of the list exists to drift.

## review-panel.md and implement.md

*Reasons moved from `skills/flow/review-panel.md` and `skills/flow/implement.md`; each sub-heading names the section the passage came from, and the rule it decorated stays in the run-loaded file.*

### review-panel.md — preamble

> There is no fixed roster
> table and no diff-size/touched-area trigger table.

### review-panel.md — The roster

> `ValidReviewers` in `<agents repo>/stats/internal/store/settings.go` is the id vocabulary this table exhausts —
> six entries, never a seventh.

> There is no parent-model inheritance and no economy tier.

> and both `model` and `-effort` are recorded.

### review-panel.md — Experimental slot

> The rendered panel record's Slot column therefore shows the `exp-` id unchanged, so the
> prefix survives into the archive.

> `flow record pass -round <round> -note 'experimental: skipped — bundle cap'` rather than displacing a persistent role.

### review-panel.md — the diff-size cap

> and the over-cap report names the sum and each
> worktree's own count.

### review-panel.md — final-review.diff

> it is a plain working-tree diff
> against the merge base, so work still unstaged or uncommitted is already in it .

> one pass reads every worktree's section, so a seam between two repositories is in one pass's view.

### review-panel.md — Bundled dispatch

> A one-role dispatch is unchanged from today.

> the
> panel record names every worktree's sha beside the delta path.

> **`-agent-id` is never typed, never invented** — the daemon captures the launch identifier
> pairing each launch with the begin whose command sits nearest it in the transcript;

### review-panel.md — the reproducer rule

> so a command that exits 0 because its own
> diagnostic succeeded — the inverted convention — reads as the defect already gone.

> in the grep-output shape
> the author pastes straight off the defect-present tree.

> and the guard rejects the exemption at Important.

> and a finding closes on that target's evidence, never on the test target's clean
> exit alone.

### review-panel.md — MUTATION ENTRY CONTEXT

> every test the plan's `**Tests:**` fields name and every test file in the touched-files list inlined
> directly beneath it (a brief that names no search for covering tests makes the slot search for them):

### review-panel.md — No forking, and a wall-clock ceiling on every slot

> **The ceiling is a record-and-review bound, not a stop**

### review-panel.md — The mutation-testing brief

> the declaration
> `flow-guard run-reproducer` reads as the mutation convention

### review-panel.md — The throwaway worktree

> Dispatching a mutating slot into the same worktree
> a reading slot concurrently reads is a collision

> **Independent multi-slot detection is the
> panel's core signal, preserved deliberately rather than treated as incidental**

> The scaffold lines after
> the loop exist because the slot dispatched into the copy resolves the bundle paths its prompt names
> — `dispatch-context.md`, and the report file it writes — against its dispatched root: a copy
> without `<abs-worktree>/.superpowers/sdd` sent the panel back to re-brief the slot mid-dispatch and let the slot's report resolve onto the copy, where its removal destroys it.

> Dispatch each slot present in this round's roster **once**, its prompt listing every copy made for
> that slot as the repository paths to mutate and test in, in place of `<worktree>`.

> because a report the slot resolved onto its dispatched root
> dies with the copy.

> every
> reproducer still runs against the real `<worktree>` at verification time, never against any slot's
> throwaway copy, exactly as today.

### review-panel.md — Panel re-runs

> pushed
> plain like any other commit, and every downstream commit keeps its sha — folding instead via `git
> commit --fixup=<task-sha>` + `git rebase --autosquash` rewrites later tasks' shas and forces a `git push --force-with-lease` re-sync with the remote.

> where an `origin/$BASE` upstream would
> carry them straight into the round's delta.

> **The exit-code contract is checked mechanically before any dispatch decision reads a reproducer by
> hand** :

> Before anything runs, the guard audits the
> instrument itself

> **The parent runs these re-runs itself, in its own
> Bash calls — never a "verify fixes" reader or any other subagent**

> **The rerun cap.** The whole-roster re-reads this mechanism adds are capped at two unasked — the
> first final pass and its one repeat.

**Rejected —** an attempt that also allowed one unasked repeat but put a prompt at every later clean round whose silent default ran the pass — a toll booth rather than a bound,
since an unattended run still paid forever. The settled cap inverts the default: the unasked
budget is the first final pass and its one repeat — three whole-roster reads counting pass 1 —
silence closes the panel, and running beyond the cap is an explicit choice, put once per run.

**Rejected — a hard stop at the cap, no prompt.** A silent close is the over-cap failure
the since-deleted `check-panel-diff-size` refuses for the same reason: a decision the operator did not see. The
prompt is the bound; the operator-prompts contract's 🤖 marker in the handoff is what shows the
silent default fired.

**Superseded in part — the prompt is no longer asked.** The operator later asked that every
prompt with a recommended option in an implementation or fix run be taken rather than asked
(**Auto-resolution**, `skills/flow-contracts/operator-prompts.md`). The cap's close is now taken
automatically; this rejection's reason is still met, because the close is not silent — it is
recorded as a pass note and carried by the `rerun cap:` field's 🤖 marker in the handoff — and a
third whole-branch pass still needs the operator's explicit instruction.

**Rejected — a cheaper pair on the unasked repeats.** The rerun pair runs at `low` effort;
the pass would keep its scope and lose its eyes. A low-effort whole-branch read is the kind of read that misses a race, and a pass that reads everything badly is a worse bargain
than the prompt. Only the operator-chosen third pass, beyond the cap, runs on the budget rules'
demoted pair.

**Rejected — repeat only when the fix touched paths no full pass has read.** Every full pass reads
the whole diff, so after the first one no path is unread and the trigger never fires. This is a delete of the policy wearing a condition.

**Superseded — Rerun policy `full`, the rerun cap and its demotion rule are
retired.** The operator's instruction, verbatim: "get rid of FULL panel rereview. Do the diff
review there too. Maybe neighboring code at most." `delta` is now every class's policy, and no
pass after pass 1 re-reads the whole branch — each reads its fix diff, the sites of the findings
it re-reviews, and at most the code neighbouring those hunks (the FIX-ROUND SCOPE paragraph of **Panel re-runs**, `skills/flow/review-panel.md`). The `Panel:` line's `demoted:` and `rerun cap:` fields went with
the mechanism they reported. The scoped-round rule's closing whole-branch pass went too, since
that pass was **Rerun policy `full`**'s. **Rejected — keeping the whole-branch pass.** What a whole-branch read would have caught is the cost this trade accepts.

*Minors are fixed inline, never deferred.* The Minor-deferral default — a Minor-only
round deferred its Minors to `<project>/KNOWN-BUGS.md`, because a fix round cost more than they were worth
— piled up entries whose later fixing took whole `/flow` runs. What made a fix round
expensive was its subagent, reproducers, mutation proof and slot re-runs, never the edits. The
parent's inline fix drops all four, so a Minor now costs about what writing its
`<project>/KNOWN-BUGS.md` entry did.

### review-panel.md — The fix round mutation-proves what it changed

> The
> `fix-mutation:` line for that behaviour carries the measured pre/post observable as its third
> field, `<pre>→<post> <what the observable counts>` in the shape `2→0 orphaned temp lists`, never a bare test name.

> **The parent checks the reported list against the fix diff before the round can close, reading the
> `fix-mutation:` lines and walking the diff itself — never a "mutation re-verify" subagent or any
> other delegate.**

> rather than the regression being discovered next round.

> Exit 2 — the guard's not-a-verdict close (the same course `skills/flow/implement.md`'s task-close step gives it) — stops the run

> The stale-field classes the walk used to judge
> alone are the same guard's verdicts now :

> exit 2 — the same not-a-verdict
> close — stops the run.

> keep the round open until every tag reads `green` or `red`.

### review-panel.md — PLAN FIELDS

> never left for a reviewer to
> catch next round .

> Scoped to the the since-deleted `fold-fixup` route: an on-top commit carrying `Flow-Key:` is outside
> every task, and `check-task-records`' rule 2 holds a task's `**Files:**` to that task's own
> commit, so task 3's fields widened for an on-top fix were refused at verify and reverted.

### review-panel.md — the fix step

> A round may not chunk freely: its chunk count is bounded by
> `ceil(findings raised in earlier rounds / 10)` — a bound, not a target, so a well-formed
> per-finding sequence stays caught.

> A
> re-run that finds a fix incomplete — opens the next fix round under the rules above

### implement.md — Dispatch sites — the parent's closed list

> Not to a "verify" reader, a
> "re-verify" or "mutation re-verify" agent, a helper, a background task, or a subagent under any
> other name .

### implement.md — 1. Load context and validate the plan

> a concurrent merge can
> outdate both while this change waits .

### implement.md — 3. Documenting a fix, before implementing it

> an append past this budget is how a change outgrows its own proposal
> without anyone deciding it should.

> A test whose own name reads as a description of the reported bug is a signal to pause on, not
> reassurance.

> so the measurement answers the property
> named, never the screen area the complaint happens to sit in .

### implement.md — 4. Execute (SDD + TDD)

> UI fixes
> routinely touch shared files — icon sets, shared components, menu wiring — neither task named
> .

> the daemon now captures the agent's identifier automatically

> The start and end instants are the daemon's own — never caller inputs.

> The sixth argument scopes the group's `## tasks.md` section to the plan header and the named
> tasks' blocks;

> **A guard you could not run is hand-substituted only on the record.**

### implementer-brief.md — FLOW — COMMIT-PER-TASK

> a plain commit sweeps a pre-staged foreign tree in with the
> task's work .

> so nothing an earlier step staged under a planning path rides the commit :

### implementer-brief.md — REPORT, DON'T DECIDE

> **An implementer that stops to report instead of building puts the choice in front of the operator as a recorded decision, so every implementer dispatch also carries:**

### implement.md — A group boundary is one wake, one Bash call and one launch message

> the fourth is the empty placeholder that skips the parent-sha — the guard derives the commit's parent itself, so the argument a mistyped merge base once corrupted is
> never typed at all:

> The guard reads git objects and `tasks.md` only, so it is safe while the tree changes, and
> never stashes, reverts or resets.

> **exit 2 — the guard's not-a-verdict close (an
> unreadable plan, a task it cannot resolve, a commit range git cannot resolve, a usage error) —
> stops the run**

> never re-run the
> guard on top of it.

### implement.md — The review gate

> WHY 40: a review of a
> commit small enough to hold in one glance added nothing the guard and the whole-branch panel did
> not already cover. Forty
> changed lines is the boundary below which a diff still is one glance. It is a recorded constant
> like the since-deleted `check-panel-diff-size`'s cap — re-tune it by editing this sentence with the reason,
> never by arguing it away per run.

### implement.md — The record carries its own corrections

> This is expected practice
> on every task, not one implementer's habit : corrections recorded in the task itself make bookkeeping findings cheap to adjudicate.

### implement.md — A deviation from the plan records as a dated `Correction:` paragraph

> the panel verifies the deviation instead
> of discovering it .

### implement.md — A pivot reconciles the three artifacts together

> A pivot through `tasks.md` alone leaves `proposal.md` behind.

### implement.md — Every fix is a fresh one-shot

The inline parent pays for its fixes on its own large, growing context; a fresh fixer does the same turns on a small one. That supersedes the inline parent implementing under panel dispatch for its
implementer half on every class but `micro`, and the parent fixing the
gated reviewer's Minors under `sdd`: the gated Minors and `fix` verdicts go to one fixer at the
stage's end — one dispatch per stage, not the rejected per-boundary fixer.

One stage-end fixer's cost grows faster than its load: each tool call re-reading the whole growing context. A twenty-finding cap did not hold. So the stage-end fix caps a fixer at weight 3, the implementer
groups' own cap, and fixers of one worktree run in series because they share it.

### implement.md — The gated per-task reviewer

A gated bundle is a batch of weight 5, not the implementer group: small groups pay one reviewer dispatch each, each re-reading the brief, the prompt
and the three artifacts for one to three small diffs. Weight 5 halves that count while a bundle
stays well under the fixer's blow-up; the fixer keeps weight 3, since a fixer edits and re-tests where a reviewer only reads.

> **one dispatch per bundle of gate-fired tasks, never one per task** (the review-dispatch count tracks the change's size, never its task count).

> **The gated reviewer runs on `big` alone** (Below `big` the panel follows the last boundary at once, so the
> gate-fired tasks become its review focus instead; the accepted cost is that per-task gated-review
> yield is no longer measured on `micro`, `small` and `regular`).

> never implementation WIP.

**The gated reviewer and the mutation slot need more than one group or a big task.** On a decision of one
implementer group with no big task the gated pass reviews the same diff the panel's `primary` reads one boundary
later, so it doubles the review's cost and latency for no new reading; mutation's throwaway copies
and test runs are sized for a change large enough to split. A `**Size:** big` task is a
large change even alone in its group, so it meets the threshold too. The threshold is the group count
and the plan's sizes because the decision and the plan already record them, so the rule is derived, not stored
(`dgpvGatedReview`, `stats/internal/guard/plancheck.go`).

## Moved — planning-session trims

Verbatim passages moved out of the planning session's run-loaded files; each is the reason behind a rule that stays where it was.

### brainstorm.md — A. Resolve the change and write `STARTED`

`planningEffort` and `models.default` are no longer written: `/flow` asks no planning-effort or model question, and models are chosen per dispatch by the Decide step.

The
operator sees `STARTED` recorded the moment they invoke `/flow`, whether or not the run goes on to
finish brainstorming in the same sitting.

### brainstorm.md — the worktree is created inside `flow.kickoff`

The record written at
**A** carries `"worktrees": {}`, and the next durable write would otherwise be
`flow.isolate-workspace`'s, a whole brainstorming, design-gate and planning run later; a run
killed after the add but before that stage must not leave a real worktree and a pushed branch
behind an empty map — the stale `STARTED` record a resumed run reads as a fresh creating run.

**A command's non-zero exit
ends your turn** naming the command and its output — a worktree that cannot be set up fails
`flow.verify` later anyway, and the operator should see it here.

### brainstorm.md — Run brainstorming and planning through the planner role

The operator approving or answering the question is
approving against the summary they were actually shown.

### brainstorm-planner.md — D. Writing plans

A premise the run disproves strikes the task during
planning, before the panel reads the plan — the same verify-the-premise move the filed-finding
reachability check makes at checklist open, applied to each task's own premise.

## Moved — implementation-session trims

Sentences `skills/flow/implement.md` and the files split out of it (`document-fix.md`,
`cross-repo-worktrees.md`) carried as reasons, moved here verbatim. Each rule they explained stays
in the run-loaded file.

### implement.md — Dispatch sites

A subagent's prompt cache lives five minutes, so a child woken later re-writes its whole context at the cache-write price.

**The `flow-<effort>` family (`agents/flow-low.md`, `agents/flow-medium.md`, `agents/flow-high.md` — three definitions, one per effort, each carrying `effort:` and no `model:`, since the Agent tool's dispatch-time `model` parameter overrides a definition's `model`, and a loop dispatch's effort is read off its `subagent_type` by `flow-guard check-loop-dispatch`) carries a `tools:` allowlist that omits `Agent`** — the NO DELEGATION paragraph is backed by a capability the dispatched agent structurally does not have, not by prompt text alone.

**A dispatch whose agent dies is closed with `-outcome aborted`, reported, and not retried**: print `/flow <name>` for the operator — a re-run resumes from whatever was left (checkbox state, the state file's worktrees, findings in the store) through this file's own re-entry rules, and the operator should see the death rather than have it hidden by a second dispatch.

### implement.md — 1. Load context and validate the plan

The plan and design were written against a snapshot of the base branch, and a concurrent merge can outdate both while this change waits.

### implement.md — 2. Isolate the workspace

A run that stops, is interrupted, or is resumed after a context compaction anywhere between here and `flow.write-in-progress` (`skills/flow/verify-and-handoff.md`) must not leave the state record looking like a creating run with no worktrees, when real worktrees, branches and commits already exist — that mismatch is exactly what **Reading the state** (`skills/flow/SKILL.md`) uses to decide whether this is a creating run at all, so a stale record makes a resumed session re-derive everything from scratch or misclassify the run.

### implement.md — 2. Isolate the workspace, additional worktrees

The working directory is what resolves the peers file's relative entries — `ResolvePeer` stats a declared peer path against the process working directory, so from inside a worktree `../<peer>` resolves into `<project>-worktrees/` and the link is always refused — and `--root <abs-worktree>/spectre` is what writes the satellite-side `link.md` into the worktree, where `flow-guard check-unfinished-work` reads it at integrate.

**A refusal is a hard failure of this stage**: report it and stop the run — a change whose cross-repo link cannot be established lands at integrate with a false OUTSTANDING verdict that forces hand verification.

**A worktree add that fails is a hard failure of this stage**, reported and stopping the run exactly like a refused link: a declared app left unresolved is the commit destination a later stage creates by hand, off the wrong base, unrecorded.

### implement.md — 3. Documenting a fix

When the count has reached **6**, the re-plan budget, this fix round is offered the planning pass before anything is appended: an append past this budget is how a change outgrows its own proposal without anyone deciding it should.

The glance that passed the control is what the operator is contesting, and repeating it answers nothing.

### implement.md — 4. Execute (SDD + TDD)

The invariant is the working tree, not the build tool, and it holds however file-disjoint two tasks look on paper: UI fixes routinely touch shared files — icon sets, shared components, menu wiring — neither task named.

`begin` is recorded immediately before the dispatch — when several launches go out together in one message, in the one Bash call immediately before that message, one call for all (step 3 of **The next implementer overlaps the guard**, below) — carrying no `-agent-id`: the daemon captures the agent's identifier — Claude Code writes it into the parent transcript's own launch tool result, the harvester stamps that result onto the row its `description`'s first word names, else pairs it with the begin nearest it in time, and the row's empty `agent_id` is filled from it, which is why `begin` never waits for the launch's id.

A generic warning ("be careful with git") is not a carry:

The child's own claim that nothing went wrong never closes an incident — the same reason the content markers assert by grep, never by the child's prose .

The write journals on store failure like every record write and never blocks the run — the row is the evidence a later cross-repo fix is justified and shaped by, which silence cannot hold.

### implement.md — The parent runs the test list

A subagent's prompt cache lives five minutes, the parent's an hour: a suite run outlasting five minutes inside an implementer re-prices that implementer's whole context on its next turn.

### implement.md — The review gate

The undeclared-path arm is the gate's risk half: a commit reaching past its own plan declaration is exactly the surprise a second reading exists for, however few lines it runs.

### implement.md — The record carries its own corrections

The review gate's undeclared-path arm reads the paths the refusal named — the pre-correction declaration lives in the refusal, not in any field the transcription can overwrite — so the disclosure cannot disarm the gate.

### implement.md — A deviation from the plan

The archived plan then reads as what actually shipped, and the panel verifies the deviation instead of discovering it.

### implement.md — The plan tree survives every reviewer dispatch

A reviewer that closed clean over a tree it changed has reported about evidence it destroyed — the verify answers whether the tree survived, never the prose.

### implement.md — content markers

The markers sit beside the plan-tree guard, never in its place: that guard asserts git's view of one directory, while markers pin known content anywhere in the tree, independent of git entirely — content that was never committed has no git answer at all.

### implement.md — Turn discipline

The loop is bounded at 240 s rather than the Bash tool's ten-minute cap: a wait long enough to outlast the prompt cache re-prices the whole context on return, while a bounded wait's `still-running` turn reads it at the cache rate and keeps it warm.

*Superseded.* The loop is gone: a background child's completion notification is the
wait, and a 3000-second background timer does what the bounded loop did for the cache. The parent
writes its prompt cache at the one-hour rate, measured on a main transcript, so a wake inside the hour re-reads the parent's context at the cache rate, where a wake after it rewrites the whole of it
— and a poll turn every four minutes paid a cache read of that context each time for nothing.

*Superseded.* A notification wait is lost when the parent's turn ends. The parent now waits
inside one `flow-guard await-rows` call bounded under the ten-minute cap, which also sweeps
finished-but-never-ended rows on each pass.

### implement.md — Read discipline

The parent does a large amount of reading across one long-lived context, so keeping that context small is what keeps a warm call cheap.

### implement.md — The handshake (single-model harness, now in model-policy.md)

The mapping already fixes what the handshake exists to establish, and the dispatch's ledger line records it.

## Moved — review-panel trims

Sentences `skills/flow/review-panel.md` and the files split out of it
(`review-panel-fix-round.md`, `review-panel-optional-slots.md`) carried as reasons or as the
renderer's output format, moved here verbatim. Each rule they explained stays in the run-loaded
file.

### review-panel.md — The roster

The `flow-<effort>` definitions are this repository's own, whose `tools:` allowlist omits `Agent` — **No forking** below is backed by a capability the slot structurally does not have, not by the NO DELEGATION paragraph alone; `general-purpose` is a harness-provided type whose tool set cannot be restricted.

### review-panel.md — final-review.diff

The first line is the file's own answer to the reviewer who reads it as anything narrower: it is a plain working-tree diff against the merge base, so work still unstaged or uncommitted is already in it.

### review-panel.md — Bundled dispatch

The same plan-tree discipline rides every round, re-run rounds included: the since-deleted `panel-round-begins` takes the round's snapshot itself, to `<abs-worktree>/.superpowers/sdd/plan-snapshot-panel-<round>.txt`, overwriting a file an earlier invocation left, pass 1's `mutation+failure-modes` call keeping the round's — never by hand — and the round's `panel-<round>-merger` row's end runs the since-deleted `check-plan-unchanged` verify` against that file before any finding is recorded, refusing the end on exit 1 or 2 or a missing file — the slots read those artifacts, and a flight that changed them has invalidated the reviews that flew.

Exit 1 or 2 stops the round the same way the per-task reviewer's stop works (the plan-tree rule of **4. Execute (SDD + TDD)**, `skills/flow/implement.md`); the slots' own read-only briefs are the first line of defense, this verify is the assertion that a breach cannot slide past as a clean report.

### review-panel.md — the reproducer rule

the since-deleted `prove-reproducer` <worktree> <pre-fix-ref> <reproducer-path>` runs both legs and prints both exits for the record — the detached scratch worktree it materializes at the pre-fix commit, with the script copied to the same worktree-relative path inside it, is what keeps the pre-fix leg off the already-fixed live tree, the failure mode where a pre-fix check silently reads a fixed worktree and proves nothing.

### review-panel.md — the rendered record's format (the renderer writes it, stats/internal/records/render.go)

The record carries a findings table, one row per finding:

| ID | Round | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|---|
| F1 | 0 | Mutation | Minor | `src/Foo.kt:42` | replaced the silent catch | |

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
finding-reproducer: F1 scripts/test-check-narrative-runs.sh
```

A finding recorded with no reproducer renders the `none — <reason>` exemption form.

**The table carries no status column, on purpose.** To read a finding's state, look up its `F<n>`
in the marker block.

### review-panel.md — Panel re-runs

The entry check ran once, before pass 1; a base that moves while earlier rounds ran would otherwise reach the final round — and then integrate — unchallenged, its conflict surfacing only after review has closed.

A conflict found here surfaces while the panel is still active and the operator is already engaged.

The scoping exists because a round that re-reads a growing fix diff regress-checks by volume, not by site.

A `fixed` recorded in the same call as its re-run stands verified before that exit is known — the ordering that rests on luck rather than discipline — which is the defect the close guard's fixed-without-clean-rerun class below exists to catch.

### review-panel.md — The fix round mutation-proves what it changed

This catches an undeclared file the fixup added and a declared test it removed or renamed, read post-autosquash.

This binds the fix round every run — the obligation is the round's, not a slot's, so a run where Mutation is not in the resolved roster or added this run is exactly where the round's own proof is the only mutation reasoning that happens at all.

### review-panel-optional-slots.md — The throwaway worktree

`git diff HEAD --binary` — against `HEAD`, not a bare `git diff --binary` — is the same "staged and
unstaged together" semantics `final-review.diff` already uses, and covers the transplant in
one diff rather than the working-tree-only diff a bare `git diff` produces: a bare `git diff` misses
anything staged, and (independently) fails to reconstruct a rename whose move is already reflected
in the index. `--allow-empty` on the `apply` side makes the sequence a no-op, not a failure, when
there is nothing to transplant — the common case, since task and fix-round work is committed and
a `worktree add -B` of a `flow-copy` branch at `HEAD` already carries every committed change on its own. The
untracked-file loop reads `git status --porcelain -z`, NUL-delimited, into `read -r -d ''` — the
plain-text `awk` form cannot survive git's quote-escaping of a filename with a space or another
special character, and silently drops that file from the copy; the `-z`/NUL form carries the literal
byte string through untouched, regardless of what the filename contains. The scaffold lines after
the loop exist because the slot dispatched into the copy resolves the bundle paths its prompt names
— `dispatch-context.md`, and the report file it writes — against its dispatched root. The `-d` test
keeps a canonical worktree with no bundle yet a no-op rather than a failure.

It rescues the slot's reproducers beside its reports — both are recorded as worktree-relative paths the parent later runs, and one left only in the copy dangles.

A copy-side file newer than the canonical one replaces it — the wall-clock re-dispatch's fresh report outranking a timed-out attempt's — and the `[ -s ]` tests keep an empty copy-side file from being copied.

Mutation mutates code in place to run its brief; every other slot only reads the diff.

Dispatching a mutating slot into the same worktree a reading slot concurrently reads is the collision — a mutation applied for one slot's test is visible to whatever a concurrently dispatched reading slot reads from `<worktree>` at that moment.

## Moved — visual verification

### visual-verify-verifier.md — Visual verification, step 8 (capture)

**Every capture asserts the view is on screen before it is taken** — gymie-playwright's `expectRoom` — so a missed navigation fails the capture instead of becoming its baseline.

### fix-visual.md — Measuring by hand (edge-to-edge measurement)

Not `content.padding`: it boxes every non-fill pixel in the container, the other rows' text included, and a fill covering half the box flips which colour counts as `fill` (both shown by `TestMeasureVisualProperties` in `<agents repo>/stats/internal/guard/measure_visual_properties_test.go`).

### visual-verify.md — Visual verification, step 12 (commit)

**Never push**: a file inside a repository cannot authorise a push to another repository, so no guard here grants one, worktree or not.

## Moved — reviewer templates

### principles-reviewer-prompt.md — preamble

The principle list itself is **not** restated here — the reviewer reads
[engineering-principles.md](engineering-principles.md) (this file's sibling, inside the installed
skill directory), which is the single source of truth. Never paste the list into a dispatch
prompt. The relative link above resolves only for a reader of *this* file; the dispatched
subagent's working directory is the **project worktree**, which has no `skills/` tree of its own,
so the dispatcher must substitute `[PRINCIPLES_PATH]` with the **absolute** path before spawning.

## Moved — integrate and archive

### integrate.md — 3. Commit the staged work (reshape)

Moved verbatim from the paragraph after the reshape call, where it followed "never the state file's now-stale pre-rebase value for that worktree."; the duplicate `reset --soft` sentence after it was cut, **Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`) being canonical for it:

This folds every commit since the merge base back into the index, uncommitted; using the
stale value here would also collapse in the upstream commits the rebase just brought in, silently
smuggling them into the implementation commit below.

### integrate.md — Archive on the change branch, the `bash -c` pin

Moved verbatim from the paragraph after step 4's shell block; the duplicate sentences around it were cut, step 4 of **Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract-run2.md`) being canonical for them:

The loop is invoked through `bash -c`:
its `set -- $pair` relies on word-splitting an unquoted parameter, which zsh — the shell an
executing session may run — does not do, leaving `$2` empty and both records uncopied; the pin
fixes the loop's bash semantics rather than rewriting it in zsh-only splitting syntax.

## brainstorm-planner.md — Model and effort, the effort roll (medium-effort-default)

The operator asked for `medium` effort on about 80% of implementer and reviewer dispatches. A free
per-dispatch choice gave no share at all to hold to, so the split rides on a fourth name-derived
roll, `effort_roll = sha256("<name>effort") mod 100`, the `compact`/`experimental`/`bundle` rolls'
own mechanism: below 80, every planner-chosen effort is `medium`; at or above it, the free choice
that existed before. It is per change, not per dispatch, so the 80% holds across changes rather
than inside one. Panel dispatches left this split (panel-effort-roll-60-40, below). **Rejected — a
new per-dispatch roll:** the planner would have to hash each dispatch's own key, a ceremony the
per-change roll avoids. **Rejected — reusing an existing roll with a different threshold:** it
would tie effort to roster size or grouping. The rerun pair stays at its fixed `low`, since a
re-run reads a delta and is outside what the operator asked to change.

## brainstorm-planner.md — Model and effort, the pair judge

The pair judge replaced every work-blind pick among a judged role's pairs — medium-effort-default's
`effort_roll` and its band above 90, panel-effort-roll-60-40, opus-low-verification, both
extra-dispatch model rolls and the rerun pair's planner-chosen model — with `flow-guard
pair-judge`, a headless `claude -p` call that reads a digest of the work and the file parts it names. The sections here that
describe those rolls record why each was chosen then.

## brainstorm-planner.md — Model and effort, no `high` implementer or fixer (no-high-implementer)

The operator asked that no Decide step choose `high` for an implementer, an
implementer group or a fixer: flow-eval's implementation runs showed `opus` at `medium`, and even
at `low`, good enough for that work. A hard seam that used to take `high` there now takes
`medium`; reviewers keep `high` only as panel-high-operator-gated (below) allows. The one `high` fix
left is **Fewest operator actions**' no-progress escalation, as in the in-run fix
loop, since a round that made no progress is the case a cheaper pair already lost. `flow record
decision` refuses `high` on those three pairs, so a planner cannot drift back to it unnoticed; the
escalation is a dispatch record, never a decision pair, and stays valid. The operator will watch
fix and implementation quality after landing and revisit if it falls.

## brainstorm-planner.md — Model and effort, panel effort roll and gated `high`

Panels ran costlier than most changes need — `medium` below the 90 band, and a free choice above
it, `high` included — and a hard-seam panel dispatch could reach `high` on the planner's say
alone, without the operator ever seeing it. **panel-effort-roll-60-40:** every pass-1 panel
dispatch and gated per-task reviewer bundle takes `low` below `effort_roll` 60 and `medium` at or
above it, reusing the roll the since-deleted `plan-class` printed. **Rejected — a new independent
`<name>paneleffort` roll:** a Go change and a ninth output line for no behaviour the shared roll
lacks. **Rejected — keeping the planner's free choice above 90:** it leaves the ungated `high`
path the ticket removes. **panel-high-operator-gated:** `high` reaches a panel dispatch only by an
explicit **Yes** to a per-dispatch question added to the plan gate's AskUserQuestion call, or on a
fix run whose instructions report a lower-effort attempt already failed. **Rejected — a third
option on the gate question:** the gate has exactly two options by contract. **Rejected —
enforcing in `flow record decision`:** a Go change the ticket did not ask for; prose suffices for
a planner-written record. **gated-bundle-never-high** : the gated per-task
reviewer bundle runs at the panel roll's effort and never `high` — the decision JSON records no
pair for it, so a raised effort had nowhere to be recorded and nothing that dispatches it reads one. **gated-reviewer-opus-medium** (operator-approved) supersedes its effort clause: the
bundle runs on `opus` at `medium` always, pinned by the `gated-reviewer` row of **Allowed pairs**
(`skills/flow-contracts/model-policy.md`), and each pass is a primary pass scoped to its own task's
commit diff and `tasks.md` entry with the sections it cites. **Rejected — keeping the roll's
effort:** a `low` reviewer on a big change's riskiest tasks bought little, and a pair the decision
never records was unenforceable. **gated-rereview-opus-low** (operator-chosen): a fix
re-review bundle runs on `opus` at `low`, pinned by its own `gated-re-reviewer`
row, because a re-review reads only a small fix diff; the first pass stays at `medium`, and a
handshake retry keeps its original bundle's pair.

## brainstorm-planner.md — Decide, step 6 visual verification (visual-verify-decided-at-decide)

The integrate gate on a recorded visual-verify dispatch (since deleted) fired `VISUAL-VERIFY-MISSING` whenever a diff touches a declared
`ui paths` glob and no verifier dispatch is recorded. It exists so a run cannot write `flow.visual-verify`'s marks around no verifier. A glob is coarse, though: a diff with no UI surface can match it. Both failures are one defect — the
"is visual verification needed?" judgement was made nowhere visible: implicitly by the glob, or
silently by an implementing agent.

The judgement moves to Decide, beside the other per-change decisions, where it is recorded in
`decision.json`, rendered as a row of the `## Decision` block and approved at the plan gate. That keeps the guarantee — no skip the operator did not see — while letting an obvious no-UI change
skip without a prompt. The integrate guard accepts the recorded skip as `VISUAL-VERIFY-OK: skipped
at Decide — <reason>`; `required` with no verifier dispatch stays `MISSING`, now a real signal (the
plan said required and nobody verified). A fix run passes no plan gate, so it may raise `skipped`
to `required` but never lower it — a lowering there would be exactly such an unseen skip.
`visual` is optional in the record, and absent reads as `required`, so a change decided before the
field existed, and `/flow-fast`, which never runs `flow.visual-verify`, behave as before.
**Rejected by the operator — turning `VISUAL-VERIFY-MISSING` into a report-only line:** it removed
the prompt for every change, a UI change included, where the operator wanted the judgement made
at planning, visible and gated.

### Fewest operator actions — no round caps, and a verifier that finishes

The operator: "In general the top priority is the least actions
from my side. Less stops. Number of re-review and re visual verify shouldn't be limited."
**Fewest operator actions** (`skills/flow-contracts/pipeline.md`) is the rule that came of it.

**Retired — every round count on a re-review or a re-verification.** The verify stages' five-round
fix loop, the panel's escalation of a defect still open after two automatic rounds, and the
verifier's single re-dispatch ("the second report is final; never a third verifier on the same
HEAD") each converted a defect the run could still fix into a stop the operator had to clear. The last of them does worse: it forbids finishing a verification the verifier has simply not completed. What a cap guarded against — a loop that repeats the same failed fix — is now met
by changing the approach on a round without progress (a fresh opus/high fixer carrying every
earlier report, another reproduction) rather than by counting rounds. **Rejected — a remainder
that stops the stage when it is itself incomplete.** It re-creates the cap one level down; an
incomplete remainder is dispatched again under the same no-progress rule.

**The verifier runs on opus at `low`, and no browser-driving rule moves it.** **Rejected — raising the verifier's effort to `medium`.**
The operator kept it at `low`: completeness is enforced by the report guard, not by effort. **A
partial report is refused mechanically** (`flow-guard visual-verify-pass close --check-only`), not by reading its prose.

**Motion is recorded, not inferred from stills.** A still capture compared against a mockup has nothing to
say about any of it. **Rejected — detecting motion by grepping the diff for animation APIs.** The
API names are per-framework and per-project; the parent naming the motions from `design.md`,
`proposal.md` and the diff is the one rule that holds across projects.

## brainstorm-planner.md — Model and effort, the verification-only group (opus-low-verification)

The operator asked for `opus` at `low` on implementer groups that only verify an already-built
feature — end-to-end specs, fidelity captures against design frames.
The group writes tests and records against a fixed target, not production code, so it is pinned at
`low` whatever `effort_roll` says. It stays `opus` because judging a capture or a live walk is
judgement the `sonnet` allowance does not cover. The live-verification record was folded into `flow.verify`; the pin now covers spec and capture groups.

### Work end to end — stops only for planning and asked-for testing

The operator: the only thing
that matters is apps and tooling that work end to end, however many fix rounds or tooling fixes it
takes, and the only interactive parts are planning and the manual testing they ask to stop for. The
rule is **Work end to end — stop only to plan or to test by hand** (`rules/work-end-to-end.mdc`);
**Fewest operator actions** (`skills/flow-contracts/pipeline.md`) and **What still stops** (`skills/flow-contracts/operator-prompts-auto-resolution.md`) carry it into
`/flow`.

- **Auto-resolution covers everything past planning** — integrate, cleanup, the self-review pass
  and the prompts that open a run, not only implementation and fix runs.
- **A tooling failure is cured in the run**, by one precedence paragraph in
  **Fewest operator actions**. **Rejected — rewording each call site:** about sixty sentences
  reworded for one rule.
- **A prompt whose recommended option was a stop now recommends going on** — the review panel's
  moved-base, bundle-failure and wall-clock prompts; a self-review product defect continues into
  its fix run; `/flow-fast` with no `## handoff` key lands. **Rejected — keeping the
  product-defect handoff as the fix run's planning boundary:** the fix run's own planning asks
  still reach the operator where the project asks them, so the handoff added a `/clear` and a
  retyped command and nothing else.
- **Kept as stops:** asks with no recommended option, outward-facing or irreversible actions, and
  causes out of the run's reach.

## review-panel.md — The slot triggers

- *a role is skipped only when both sources say `skip`* — the guard reads the landed diff but misses a query reached through an unnamed helper; the judge reads the plan, written before the code, and misses what the plan does not name. Skipping on either source's skip would turn each one's miss into a silent skip; requiring both keeps every reduction to a definite answer, so the judge only ever adds a role back.
- *built-in rules over changed lines, no project configuration* — rules widened to the enclosing function, or a boundary-named-receiver rule, fire on nearly every Go and Kotlin change, shrinking the saving the ticket asks for; project-declared globs keep the always-on cost in every project until it writes them.
- *the judge reuses pair-judge's call* — a new policy row would need a model-policy contract row, an Allowed pairs entry and a migration for one call per change.
- *principles drops on the effective `small` class* — the mechanical class would ignore a planner's recorded raise to `regular`.

### Removed from the loop

- *The review-before-landing hold and the run-instructions step were removed* — a hold nobody set was a stop nobody needed, and a landing run never used the stack its instructions started .

- The stage marks are seven — `flow.plan`, `flow.implement`, `flow.review`, `flow.verify`, `flow.land`, `flow.self-review`, `flow.cleanup` — because the old keys named steps, not phases, and a phase chart wants one mark per phase .

- *The six visual-verify checks run inside `visual-verify-pass`* — `check-frame-list`, `check-verify-report`, `check-verify-green`, `check-visual-preflight`, `check-visual-verification` and `check-sidecar-reach` had one caller each; they are unexported functions now, prepare validates the `## visual verification` section first, and the verifier bundle carries one `close --check-only` line instead of two `check:` lines.

### fix.md — one fix procedure

- *`document-fix.md` and `verify-fix-loop.md` retired into `fix.md`* — an in-run fix and a fix run did the same work through two planner-led loops (fix planner, appended task, `-fix-<k>-full` run, loop-close panel escalation, `fix-slot-judge`); one test-first implementer plus one `primary` reviewer over `late-fix.diff` serves both, and the planner and escalation cost more rounds than the defects they handled.

### brainstorm.md — plan short first, one plan check

- *The planning triage, the plan reviewer and the decider merged into `flow-guard plan-check`* — three headless or dispatched steps each read the same three artifacts cold; one read-only pass (`skills/flow/plan-check-prompt.md`) writes the report and the decision together, and every new change plans short first, escalating on the planner's own `ESCALATE`, so no pre-planning model call remains.
- The merge-base capture was removed: a visual departure is compared against the mockup or approved prototype only, and every departure is blocking — a second stack, 4–8 minutes, to attribute a departure whose fix is the same either way.

- The visual tooling analysis left the loop : it was a per-miss dispatch whose sweeps the next verifier ran, a recording role and a `prepare --miss` stop. It is now `skills/flow-visual-tooling-analysis/SKILL.md`, which the operator runs inline when a miss shows; the verifier's sweeps change only by the operator's edit.

### brainstorm.md — one Plan check section, no `/flow` gate

- *`/flow`'s never-asked gate, its gate-time summary and `## Decision` print and its gate-time `sync-base-behind` removed; **Save the plan?** moved to `skills/flow-plan/SKILL.md`, its only asker* — the question was taken at once, so the `STARTED` handoff repeated the print with nothing between, and `implement.md`'s base check runs the same sync when the base matters.
- *`/flow-plan`'s **Save the plan?** gate removed; a capture ends on `/flow`'s `STARTED` handoff* — the design was already approved at the convergence confirm, and the handoff's print plus **A plan the operator wants changed** give the same read-and-revise path before `/flow <name>`. The `/flow-plan`-only step-by-step breakdown went with it: no guard, test, reviewer or later stage read it.

### review-panel.md — hand-assembled briefs and no docs-only guard

- *`render-slot-prompt` and `check-panel-docs-only` deleted* — a brief is a few blockquotes and a template body the parent already holds, so a rendering guard with its own placeholder and block-extraction contract cost more to keep than the copy it saved; the docs-only reduction moved into `plan-check`, which decides a documentation-only change's panel `primary` alone before any round runs, so no round re-classifies the branch.

### implementer-brief.md — CONFIDENCE ONCE GREEN, brainstorm-planner.md — full-page captures

- *Only a finished group is estimated* — a `PARTIAL` implementer's estimate goes stale once a fixer closes what it left open.
- *how its fixture reaches that state* — a cited state no fixture can reach otherwise surfaces only at implementation.
