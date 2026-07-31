## Context

The full brainstorming record for this change is
`docs/superpowers/specs/2026-07-30-kan-17-finish-gate-jira-and-commit-hygiene-design.md`. This
document carries the technical decisions; that one carries the dialogue they came from.

KAN-17 lists eighteen asks. Brainstorming split them into two changes. This one takes the eleven
that change how the pipeline behaves, plus one ask the operator added during the planning run
itself: `/myflow-start` must put every unresolved question to the operator rather than resolve it by
assumption. The remainder — the asks that change what the operator sees and how the pipeline is
configured — go to a follow-up issue filed at the end of this planning run.

Three of the surfaces involved have no capability in `openspec/specs/` today: Jira, the review
panel's findings bar, and `/myflow-start`'s conduct. That is the common thread. The In Review defect
survived four archived changes precisely because no requirement existed for it to violate.

## Goals / Non-Goals

**Goals:**

- Make Jira behaviour normative, and correct the two defects the missing spec allowed.
- Refuse to integrate silently over unfinished work, and record it when the operator chooses to.
- Give review findings a machine-readable status, so "every severity blocks" is checkable.
- Take planning artifacts out of the diff a human reviews, rather than filtering them at display.
- Put every temporary artifact in one registry, and close the remote-branch gap.

**Non-Goals:**

- The seven deferred asks, tracked separately.
- Renaming `myflow-review-panel-economics` to match its widened subject — a rename would pull
  `myflow-contract-distribution` into a change that is not about distribution.
- Adding a test harness for `check-vocabulary.sh`, the one guard script that lacks one.
- Any change to the three states, the panel roster, the model policy, plan provenance, or the state
  file's schema.

## Decisions

### The issue moves to In Progress when planning begins

**ID:** `jira-in-progress-at-start`
**Status:** active
**Chosen:** Transition immediately after the Jira key resolves, before the first brainstorming
question — the board is then correct while a long planning run is in flight.
**Considered:** Keep it after the state write and report the before-and-after status more loudly —
rejected as not fixing the board during the run. Treat the ask as fully subsumed by
`jira-unknown-status-ask` — rejected; the operator wanted the earlier timing on its own merits.

Non-blocking is preserved, and that is the load-bearing point. The rule this replaces existed so a
Jira failure could never prevent the state from being recorded; moving the call earlier does not
reintroduce blocking, because the degradation path is unchanged — a failure is one
`⚠ Jira: skipped — <reason>` line and the run continues to write state normally.

### In Review fires on every landing route

**ID:** `jira-in-review-any-route`
**Status:** active
**Chosen:** At the end of finish run 1, whichever route was taken — route-independence is the fix;
conditioning on a PR is what broke it.
**Considered:** Keep the PR condition and add the manual route — rejected, because merge-and-push
would still skip In Review and the same class of gap survives. Fire at the start of run 1 —
rejected: it would claim In Review even when the push then fails and the run stops.

### An unrecognised status is asked about, never inferred

**ID:** `jira-unknown-status-ask`
**Status:** superseded by `jira-unknown-status-ask-bounded`
**Chosen:** When the current status is not one of the four ordered names, show the key, the current
status and the target, and ask whether to transition.
**Considered:** Order by Jira's `statusCategory` — rejected, and this is the important rejection:
`TO DO URGENT` and `In Progress` share the `indeterminate` category, so category ordering reports
the issue as already at the target and the board freezes for the whole change. Declare the order in
`.myflow/project.md` — rejected as a second list that misorders silently when stale. Treat unknown
as before In Progress and always transition forward — rejected: it would drag a status such as
`Blocked` backwards.

### Created issues inherit the parent's labels plus `AI-generated`

**ID:** `jira-created-issue-labels`
**Status:** active
**Chosen:** The parent issue's labels, plus `AI-generated`; no label is ever invented.
**Considered:** `AI-generated` alone — rejected; follow-ups lose the routing label that makes them
findable. A label list in `.myflow/project.md` — rejected as another list to maintain.

"Use only existing ones" needs no enumeration step: the parent's labels exist by construction, and
`AI-generated` is already in use in this project.

### The run-1 gate is computed by a script, not asserted in prose

**ID:** `unfinished-work-script`
**Status:** active
**Chosen:** `scripts/check-unfinished-work.sh` prints one verdict line, in the shape
`check-finish-preflight.sh` established.
**Considered:** Prose contract only — rejected: prose is the enforcement class that already failed,
in the kan-6 case this ask exists to prevent. Script plus a `## lint` entry — rejected: the guard
would fire on every change in flight, which is the normal mid-work state.

### Four signals count as unfinished

**ID:** `unfinished-work-signals`
**Status:** active
**Chosen:** Unticked boxes in the manual-test guide; unchecked items in `tasks.md` and any nested
`<name>-fix-N`; findings not closed in the panel record; and a `## Known incomplete` section.
**Considered:** The guide's boxes alone, which is the literal ask — widened by the operator.

### Undone work is recorded in the manual-test guide

**ID:** `known-incomplete-location`
**Status:** active
**Chosen:** A `## Known incomplete` section in `docs/manual-test/<name>.md`, written by
`/myflow-do` and refreshed on every fix run.
**Considered:** The review-panel record — rejected as an audit file the operator does not normally
open. A state-file field — rejected: the state file carries only fields with a live consumer and
records no human judgment. Appending to `proposal.md` — rejected as buried in an artifact read once.

Finish runs in a different session from `/myflow-do` and has no memory of it, so this signal needs a
durable home or it cannot exist at all.

### The gate offers three answers

**ID:** `unfinished-work-prompt`
**Status:** superseded by `unfinished-work-prompt-recommended`
**Chosen:** Continue and integrate anyway; stop so the operator can finish it; or file a Jira task
and continue.
**Considered:** A fourth answer handing back to `/myflow-do` inline and resuming — rejected as
making one command re-enter another mid-run.

### Continuing past the gate is recorded durably

**ID:** `unfinished-work-audit`
**Status:** active
**Chosen:** The outstanding list appears in run 1's handoff and in the planning commit's message,
so the git history answers whether anything was known-unfinished at the merge.
**Considered:** Handoff only — rejected; the record dies with the scrollback, which is how the kan-6
case became unprovable. A Jira comment as well — not taken, as a third write to keep consistent for
no durability the commit message does not already give.

### Findings become countable, and every severity blocks

**ID:** `findings-table-and-bar`
**Status:** superseded by `findings-table-normative-format`
**Chosen:** The panel record gains a required table — slot, severity, location, status, note —
whose status is `open`, `fixed` or `withdrawn`, and `/myflow-do` may not hand off with any row open.
**Considered:** A checkbox per finding — rejected: severity and slot stay in prose, so nothing could
report which slot raised what. No format at all, with the agent asserting the bar is met — rejected;
an unverifiable claim is how minors were quietly deferred before.

Where a disputed minor ends: the panel's existing ladder hands back to the operator at the last fix
round, and `withdrawn` is written there, by the operator, with a reason. That is the existing human
gate rather than a new escape hatch, and it is stated rather than left to be discovered.

### The panel's findings bar joins the existing panel capability

**ID:** `findings-spec-home`
**Status:** active
**Chosen:** `myflow-review-panel-economics`, giving one capability for what the panel costs and what
it must produce.
**Considered:** A new `myflow-review-panel-findings` capability — rejected: two specs about one
panel makes "which do I amend?" a question. `myflow-command-surface` — rejected as being about the
command surface, not review quality.

### Planning artifacts leave the staged diff entirely

**ID:** `planning-artifacts-unstaged`
**Status:** active
**Chosen:** `/myflow-do` stages everything except `openspec/`, `docs/manual-test/` and
`docs/superpowers/`; finish run 1 commits implementation first and those three paths second; a fix
run with a PR open makes the same two commits.
**Considered:** `openspec/` only, the literal ask — rejected: the test guide and preserved records
would stay mixed into the reviewed diff. `openspec/` plus `docs/superpowers/`, leaving the guide
with the code — rejected by the operator. Planning commit first — rejected: the newest commit on the
branch would then be a docs commit, which is the one a forge shows first. Splitting at finish only —
rejected: the split would hold for the first integration and not for any fix pushed to the PR.

This converts `myflow-handoff-output`'s existing requirement from a display filter into a staging
rule, and its scenarios change with it.

### The exclusion list is fixed in the contract

**ID:** `exclusion-list-fixed`
**Status:** active
**Chosen:** `openspec/`, `docs/manual-test/`, `docs/superpowers/`, named in the contract.
**Considered:** A `## review-excludes` key in `.myflow/project.md` — rejected: the pipeline itself
chooses these paths, so no project can currently differ, and an unset key would need the fixed list
as its default anyway.

### Cleanup becomes one registry, and the remote branch joins it

**ID:** `cleanup-registry`
**Status:** active
**Chosen:** One table in `pipeline.md` listing every artifact the pipeline creates, with what
creates it, where it lives and what removes it — plus automatic deletion of `origin/openspec/<name>`
at run 2, treating an already-absent branch as success.
**Considered:** Asking before the remote delete — rejected: run 2 has already proved the branch is
an ancestor of the base branch, so its commits are in the base branch and nothing is lost. Gating
the remote delete on the local `git branch -d` succeeding — rejected as leaving the remote branch
behind whenever anything unrelated failed.

### Run 2 verifies the registry with its own script

**ID:** `cleanup-verification-script`
**Status:** superseded by `cleanup-verdict-gates-finished`
**Chosen:** `scripts/check-cleanup-complete.sh`, returning a verdict the way the run-1 gate does.
**Considered:** Prose reusing the existing per-removal verification — rejected by the operator in
favour of mechanical enforcement. Folding it into `check-unfinished-work.sh` — rejected: it is a
different question at a different time, and no script in this repository answers two.

### `/myflow-start` asks rather than assumes, everywhere approval is sought

**ID:** `planning-gate-always-ask`
**Status:** active
**Chosen:** A new `myflow-planning-gate` capability holding two rules — every unresolved question is
put to the operator rather than resolved by assumption, and every approval or choice a `/myflow-*`
command asks for is offered as options rather than open prose.
**Considered:** Adding it to `myflow-effort` — rejected: that spec's subject is the effort level.
Contract text with no spec — rejected as reproducing the unspecced shape this change fixes
elsewhere. Scoping the options rule to one command — rejected in favour of one rule covering every
command, which subsumes both readings of the ask and costs a single sentence.

### The unrecognised-status ask is a bounded carve-out from never-blocking

**ID:** `jira-unknown-status-ask-bounded`
**Status:** active
**Supersedes:** `jira-unknown-status-ask`
**Chosen:** The ask stands, but it is now an explicit, bounded carve-out from **Never blocking**
rather than an unexamined exception to it.
**Why it changed:** the review panel's robustness lens found that this change made two decisions
that were never reconciled with each other. Moving the transition to the front of `/myflow-start`
put a mandatory interactive ask ahead of brainstorming, while that skill's unmodified guardrail
says a Jira call may never "block, delay, or alter the proposal". `design.md` had argued
non-blocking purely in terms of call *failure*. An ask that waits on a yes/no delays by definition.

### The gate's three courses carry a marked recommendation

**ID:** `unfinished-work-prompt-recommended`
**Status:** active
**Supersedes:** `unfinished-work-prompt`
**Chosen:** The same three courses, with **Stop** marked as the recommendation.
**Why it changed:** the panel's primary reviewer found the prompt was a literal gap against
`myflow-planning-gate`'s SHALL that a recommended option be identified rather than left for the
operator to infer. The original decision recorded the three courses without settling that question.
**Stop** is the recommendation because it is the only course that loses nothing: the other two
integrate over work the run has just reported as unfinished.

### The emitting side writes a marker line, and the guard counts markers

**ID:** `findings-marker-lines`
**Status:** active
**Supersedes:** `findings-table-normative-format`
**Chosen:** `/myflow-do`'s panel writes, beside every human table row, one anchored line
`finding-status: F<n> <status>`, plus one `findings-total: <n>` checksum. `check-unfinished-work.sh`
counts those lines with anchored patterns and **does not parse the findings table at all** — the
cell splitting, the escaping rule, the header match, the table-boundary state machine, the orphan
net and the unparseable accounting are deleted. What remains of the table's shape is one anchored
prefix, `| F<n> |`, used only to hold the row identifiers against the marker identifiers, and every
way of getting it wrong makes the two lists differ, which is outstanding.
**Considered:** patching the parser a fourth time — rejected: each of the previous three fixes
closed the shape it was shown and left the class alive. Counting the human rows by re-deriving the
table's boundaries — rejected as the deleted defect class returning under another name.
**Why it changed:** the parser failed **open** six distinct ways across three review passes — a
capitalised `Open`; `open (needs discussion)`; an unescaped `|` in an earlier cell; a reordered
header; a row missing its LEADING pipe, which ended table tracking and hid every row below it; and a
row both detached and malformed, which each half of the parser left to the other. A findings table
written inside a blockquote was invisible once a legitimate table existed elsewhere, and awk's `==`
compares through the locale's collating sequence, so a zero-width character in a status reported
outstanding under UTF-8 and vanished under `LC_ALL=C`. The panel's simplicity lens diagnosed the
shape rather than the code: the guard was recovering one fact from a document grammar defined in
another file's prose. This repository had already paid for that lesson once — `.myflow/project.md`
records `check-plan-provenance`'s Bash classifier taking five review passes and seven fix waves
before it was replaced with a real parser — and `check-cleanup-complete.sh`, in this same change,
already solved the analogous problem the right way with `registry-row-checked:` markers.

### The `## Known incomplete` fence scan implements CommonMark's fence rule

**ID:** `known-incomplete-commonmark-fences`
**Status:** active
**Chosen:** The section scan tracks the opening marker's character and length, requires a closing
marker of the same character and at least that length carrying nothing but whitespace, and applies
CommonMark's at-most-three-spaces indent rule. An unclosed fence runs to end of file; a backtick
opener whose info string contains a backtick opens nothing.
**Why it changed:** the previous boolean toggle desynchronised on the first marker that is not a
real closer, after which every line was read with the fence state inverted. Two reproductions: a
closing marker indented four spaces inside an unclosed fence, and a ```` fence documenting a ```
one. Each made a guide with **no** real `## Known incomplete` section report `CLEAR:` — silent
clearance, which is the failure the whole guard exists to break. A previous fixer declined the
partial fix on the grounds that the indent rule alone closes only the first; the panel's adversarial
slot judged that insufficient justification to ship, and the full rule is what shipped.

### The findings table's format is normative, and an unparseable row is outstanding

**ID:** `findings-table-normative-format`
**Status:** superseded by `findings-marker-lines`
**Supersedes:** `findings-table-and-bar`
**Chosen:** The table's column order and its pipe escaping are SHALLs in the spec; the status is
compared for equality, case-insensitively, against the three legal values; and **anything
unrecognised or unparseable counts as outstanding**. A `withdrawn` row with an empty note does not
close a finding.
**Why it changed:** three panel slots independently demonstrated that the original positional regex
failed **open** — `Open` capitalised, `open (needs discussion)`, an unescaped `|` anywhere in an
earlier cell, or a column reorder each hid a live Critical finding and yielded `CLEAR:`. The
mechanism this change substitutes for prose enforcement had inherited the same silent-failure mode
one layer down. The corrected parser immediately caught four rows of this change's own panel record
carrying unescaped pipes.

### A `LEFTOVER:` verdict gates the `FINISHED` write

**ID:** `cleanup-verdict-gates-finished`
**Status:** active
**Supersedes:** `cleanup-verification-script`
**Chosen:** The script stands, and its verdict is now load-bearing: `LEFTOVER:`, or no verdict at
all, leaves the change at `IN_PROGRESS` — still listed by `/myflow-status`, still re-runnable — so
the state file it already has is the durable record. No state-file field was added; the non-goals
exclude schema changes.
**Considered:** A new state-file field recording cleanup verification — rejected as a schema change
this change explicitly excluded, and unnecessary once the verdict gates the write.
**Why it changed:** two panel slots found that step 7 wrote `FINISHED` regardless of the verdict,
after which `/myflow-finish` refuses to run and `/myflow-status` omits the change — so the only
record of a leftover was one console line. This change's own spec says a record existing only in
the session transcript does not satisfy the durability requirement. It was enforcing that principle
for run 1 and violating it for its own run-2 feature.

## Implementation notes

### The staging pathspec

```bash verified:run in a scratch repository on this machine; output reproduced below
git -C "$WT" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```

With `src/a.kt`, `openspec/changes/x/tasks.md` and `docs/manual-test/x.md` present in a fresh
repository, that command stages `src/a.kt` alone and leaves both excluded paths reported by
`git ls-files --others --exclude-standard`.
<!-- measured: the scratch-repo check above, re-runnable; git version 2.50.1 (Apple Git-155) on this machine -->

The handoff's review command therefore simplifies to `git -C <worktree> diff --cached`, with nothing
left to filter out.

`preserve-session-records.sh` keeps its existing call position in both commands — before the first
`git add -A` in finish, after staging in `/myflow-do` — because `docs/superpowers/` is one of the
excluded paths and is picked up by the second staging pass either way.

### The run-1 gate's inputs

| Signal | Read from |
|--------|-----------|
| Unticked checklist boxes | `docs/manual-test/<name>.md` |
| Unchecked plan items | `openspec/changes/<name>/tasks.md` and any nested `<name>-fix-N` |
| Findings not closed | `.superpowers/sdd/final-review-panel.md` |
| Work the run knows is undone | the `## Known incomplete` section of the test guide |

Verdicts follow `check-finish-preflight.sh`'s contract: `CLEAR`, `OUTSTANDING` with a per-signal
breakdown, and exit code two with no verdict line when the worktree cannot be read — which is
treated as stop-and-ask rather than as either verdict, because a caller grepping for `CLEAR` in
empty output finds nothing.

A missing file, or a missing `## Known incomplete` section, counts as outstanding rather than clear.
Silence is what let the kan-6 case through.

### The cleanup registry

| Artifact | Created by | Location | Removed by |
|----------|-----------|----------|-----------|
| Per-task and review diffs | `/myflow-do` | `.superpowers/sdd/` in the worktree | with the worktree, at run 2 |
| Panel record | `/myflow-do` | `.superpowers/sdd/` | preserved at run 1; removed with the worktree |
| SDD ledger | `/myflow-do` | `.superpowers/sdd/` | preserved at run 1; removed with the worktree |
| Proposal artifact source | `/myflow-start` | the state directory | run 2, only if a preserved copy exists |
| Worktree | `/myflow-do` | per the `worktrees` keys | run 2, after its existing checks |
| Local branch | `/myflow-do` | the repository | run 2, `git branch -d` |
| Remote branch | finish run 1 | `origin` | run 2 — new in this change |
| Change directory | `/myflow-start` | `openspec/changes/<name>/` | moved to the archive, never deleted |
| State file | every command | the state directory | never — it is the terminal record |

## Risks / Trade-offs

- **An abandoned planning run leaves its issue In Progress** → accepted with
  `jira-in-progress-at-start`; the board being correct during a long run was judged worth more, and
  the operator can move it back by hand.
- **Every severity blocking can cost fix rounds on matters of taste** → bounded by the panel's
  existing handback to the operator at the last fix round, where a finding can be marked
  `withdrawn` with a reason.
- **Guides predating this change have no `## Known incomplete` section**, so their next finish
  prompts once → correct behaviour and self-clearing; the alternative is treating silence as
  clearance, which is the failure being fixed.
- **Two scripts and two harnesses is the heavier option** for cleanup verification → chosen
  deliberately for mechanical enforcement over prose.
- **Unstaged planning artifacts sit untracked in the worktree between `/myflow-do` and finish** →
  finish run 1 stages them with `git add -A` before its second commit, and run 2's existing
  untracked-file check would refuse to remove a worktree that still held them.

## Open Questions

None. Every question raised during brainstorming was answered by the operator and is recorded as a
decision above.
