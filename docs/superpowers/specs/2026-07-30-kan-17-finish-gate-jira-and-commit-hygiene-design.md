# KAN-17 — finish gate, Jira projection, and commit hygiene

**Issue:** KAN-17 "Myflow updates" · **Change:** `kan-17-finish-gate-jira-and-commit-hygiene`
**Planning effort:** medium · **Date:** 2026-07-30

## Scope

KAN-17 carries eighteen asks. Brainstorming split them into two changes; this design covers
**slice A**, the eleven that change how the pipeline behaves, plus one ask added during planning.

| # | Ask | In this change |
|---|-----|----------------|
| 1 | The issue never reaches In Review | yes |
| 2 | In Progress timing | yes |
| 4 | `openspec/` out of the reviewed diff, separate commit at finish | yes |
| 5 | Always fix minor findings from the panel | yes |
| 7 | Never merge with unfinished work (the kan-6 case) | yes |
| 8 | Unticked test-guide boxes ask before integrating | yes |
| 9 | Label issues created during a myflow run | yes |
| 12 | Unfinished work is finished, or filed as a task | yes |
| 13 | Approval prompts offer options, not free prose | yes |
| 17 | `AI-generated` on issues created via the CLI | yes |
| 18 | Standardise cleanup of temporary files and branches | yes |
| **19** | **`/myflow-start` always asks rather than assuming** | **yes — added during this planning run** |

Ask 13 moved into slice A during brainstorming: it and ask 19 are one requirement, and splitting
them would make slice B amend a spec slice A had just written.

**Deferred to slice B** — a follow-up issue filed at the end of this run, carrying these verbatim:
ask 3 (`/myflow-status` reproduces the full next-step handoff), ask 6 (terminal tab colour and
rename), ask 10 (README diagram as a detailed per-command flow), ask 11 (`planning effort`, levels
`default`/`detailed`/`low`), ask 14 (the task-progress view), ask 15 (model choices recorded in the
state file), ask 16 (a simpler, wider manual-test guide).

## Why these twelve

**Ask 1 is a confirmed defect, and its cause is in the contract rather than in a slip.** KAN-20's
changelog records two status transitions and no third: `To Do → In Progress`, then
`In Progress → Done`. `jira-integration.md` transitions to In Review *"once a PR is confirmed
open"*, and KAN-20 landed by the merge-and-push route, so no PR ever existed and the condition
never fired.

**Ask 2 is not what it appears.** KAN-20 *did* move to In Progress during its `/myflow-start` run.
The issues that did not move — KAN-13 and KAN-17 — are both in `TO DO URGENT`, a status the
forward-only ordering (`To Do → In Progress → In Review → Done`, matched by name) does not
recognise, so "is it already at or past the target?" has no defined answer. The operator's decision
was nonetheless to move the transition to the *start* of the run, for board accuracy during long
planning sessions.

**Three of the surfaces this change touches have no normative source at all.** There is no Jira
capability in `openspec/specs/`, no spec for `/myflow-start`'s conduct, and the panel's
"zero open findings" bar lives only in `myflow-do/SKILL.md`'s guardrails. Unspecced behaviour is
what let the In Review defect sit unnoticed across four archived changes.

**Ask 7 names a real loss.** The kan-6 branch merged carrying an unfixed defect and five unticked
test-guide boxes. That is only knowable today because it was typed into a chat transcript; nothing
in the repository records it.

## Decisions

### The issue moves to In Progress when planning begins

**ID:** `jira-in-progress-at-start`
**Status:** active
**Chosen:** Transition immediately after the Jira key resolves, before the first brainstorming
question — for board accuracy while a long planning run is in flight.
**Considered:** Keep it after the state write and report the before→after status more loudly —
rejected as not fixing the board during the run. Treat ask 2 as fully subsumed by
`jira-unknown-status-ask` — rejected; the operator wanted the earlier timing on its own merits.

**Non-blocking is preserved, and that is the load-bearing point.** The rule this replaces
("transition after the state write") existed so a Jira failure could never prevent the state from
being recorded. Moving the call earlier does not reintroduce blocking, because the degradation path
is unchanged: a failure is one `⚠ Jira: skipped — <reason>` line and the run continues to write
state normally. **Accepted trade-off:** an abandoned planning session leaves the issue In Progress
with no change to show for it.

### In Review fires on every landing route

**ID:** `jira-in-review-any-route`
**Status:** active
**Chosen:** At the end of finish run 1, whichever route was taken — PR, merge-and-push, or manual.
Route-independence is the fix; conditioning on a PR is what broke it.
**Considered:** Keep the PR condition and add the manual route — rejected, because merge-and-push
would still skip In Review and the same class of gap survives. Fire at the start of run 1 —
rejected: it would claim In Review even when the push then fails and the run stops.

A run that stops on a failed push does **not** transition. "End of run 1" means after the chosen
route completed.

### An unrecognised status is asked about, never inferred

**ID:** `jira-unknown-status-ask`
**Status:** active
**Chosen:** When the current status is not one of the four ordered names, show the key, the current
status and the target, and ask whether to transition.
**Considered:** Order by Jira's `statusCategory` — rejected, and this is the important rejection:
`TO DO URGENT` and `In Progress` share the `indeterminate` category, so category ordering reports
the issue as already at the target and the board freezes at `TO DO URGENT` for the whole change.
Declare the order in `.myflow/project.md` — rejected as a second list to keep in step, which
misorders silently when stale. Treat unknown as before In Progress and always transition —
rejected: it would drag a status like `Blocked` or `Ready for QA` backwards.

### Created issues inherit the parent's labels plus `AI-generated`

**ID:** `jira-created-issue-labels`
**Status:** active
**Chosen:** The parent issue's labels, plus `AI-generated`. No label is ever invented.
**Considered:** `AI-generated` alone — rejected; follow-ups lose the routing label that makes them
findable. A label list in `.myflow/project.md` — rejected as another list to maintain.

"Use only existing ones" needs no enumeration step: the parent's labels exist by construction, and
`AI-generated` already exists in this project — KAN-25 carries it.

### The run-1 gate is computed by a script, not asserted in prose

**ID:** `unfinished-work-script`
**Status:** active
**Chosen:** `scripts/check-unfinished-work.sh` prints one verdict line, in the shape
`check-finish-preflight.sh` already established.
**Considered:** Prose contract only — rejected: prose is the enforcement class that already failed
once, in the kan-6 case this ask exists to prevent. Script plus a `## lint` entry — rejected: the
guard would fire on every change in flight, which is the normal mid-work state, and a guard that
cries wolf gets ignored.

### Four signals count as unfinished

**ID:** `unfinished-work-signals`
**Status:** active
**Chosen:** Unticked boxes in the manual-test guide; unchecked items in `tasks.md` and any nested
`<name>-fix-N`; findings not closed in the panel record; and a `## Known incomplete` section in the
guide.
**Considered:** The guide's boxes alone, which is the literal ask — widened by the operator to all
four.

### Undone work is recorded in the manual-test guide

**ID:** `known-incomplete-location`
**Status:** active
**Chosen:** A `## Known incomplete` section in `docs/manual-test/<name>.md`, written by
`/myflow-do` and refreshed on every fix run.
**Considered:** The review-panel record — rejected as an audit file the operator does not normally
open. A state-file field — rejected: the state file's own contract says it carries only fields with
a live consumer and records no human judgment. Appending to `proposal.md` — rejected as buried in
an artifact read once at `STARTED`.

Finish runs in a different session from `/myflow-do` and has no memory of it, so the fourth signal
needs a durable home or it cannot exist at all.

### The gate offers three answers

**ID:** `unfinished-work-prompt`
**Status:** active
**Chosen:** **Continue — integrate anyway**, **Stop — I'll finish it first**, **File a Jira task,
then continue**.
**Considered:** A fourth option that hands back to `/myflow-do` inline and resumes — rejected by the
operator as making one command re-enter another mid-run.

### Continuing past the gate is recorded durably

**ID:** `unfinished-work-audit`
**Status:** active
**Chosen:** The outstanding list appears in run 1's handoff **and** in the planning commit's
message, so `git log` answers "was anything known-unfinished when this merged?"
**Considered:** Handoff only — rejected; the record dies with the scrollback, which is exactly how
the kan-6 case became unprovable. Also commenting on the Jira issue — not taken, as a third write
to keep consistent for no extra durability over the commit message.

### Findings become countable, and every severity blocks

**ID:** `findings-table-and-bar`
**Status:** active
**Chosen:** The panel record gains a required table — slot, severity, location, status, note —
whose status is `open`, `fixed` or `withdrawn`. `/myflow-do` may not hand off with any row `open`,
at any severity.
**Considered:** A checkbox per finding — rejected: severity and slot stay in prose, so nothing can
report "two open Minors from Bugbot". No format, with the agent asserting the bar is met — rejected;
an unverifiable claim is how minors were quietly deferred before.

**Where a disputed minor ends.** The panel's existing ladder hands back to the operator at fix
round 5. `withdrawn` is written there, by the operator, with a reason. That is the existing human
gate rather than a new escape hatch — but it is the one place a minor can end up unfixed, and it is
stated here rather than left to be discovered.

### The panel's findings bar joins the existing panel capability

**ID:** `findings-spec-home`
**Status:** active
**Chosen:** `myflow-review-panel-economics`, giving one capability for what the panel costs and
what it must produce.
**Considered:** A new `myflow-review-panel-findings` capability — rejected: two specs about one
panel makes "which do I amend?" a question. `myflow-command-surface` — rejected as being about the
command surface, not review quality.

The chosen spec's name skews toward economics. That is accepted rather than renamed here; a
capability rename is a distribution concern that would pull `myflow-contract-distribution` into a
change that is not about it.

### Planning artifacts leave the staged diff entirely

**ID:** `planning-artifacts-unstaged`
**Status:** active
**Chosen:** `/myflow-do` stages everything except `openspec/`, `docs/manual-test/` and
`docs/superpowers/`; finish run 1 commits code first, then those three paths in a second commit; a
fix run with a PR open makes the same two commits.
**Considered:** `openspec/` only, the literal ask — rejected: the test guide and preserved records
would stay mixed into the reviewed diff. `openspec/` plus `docs/superpowers/`, leaving the guide
with the code — rejected by the operator. Planning commit first — rejected: the newest commit on
the branch would be a docs commit, which is the one a forge shows first. Splitting at finish only —
rejected: the split would then hold for the first integration and not for any fix pushed to the PR.

This converts `myflow-handoff-output`'s existing requirement from a **display filter** (the review
command excludes `openspec/`) into a **staging rule** (those paths are never staged before finish).
The requirement's scenarios change with it.

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
creates it, where it lives, its lifetime and what removes it — plus automatic deletion of
`origin/openspec/<name>` at run 2, treating "already gone" as success.
**Considered:** Asking before the remote delete — rejected: run 2 has already proved the branch is
an ancestor of the base branch, so its commits are in the base branch and nothing is lost. Gating
the remote delete on the local `git branch -d` succeeding — rejected as leaving the remote branch
behind whenever anything unrelated failed.

### Run 2 verifies the registry with its own script

**ID:** `cleanup-verification-script`
**Status:** active
**Chosen:** `scripts/check-cleanup-complete.sh`, returning a verdict the way the run-1 gate does.
**Considered:** Prose reusing the existing per-removal verification — the recommendation, rejected
by the operator in favour of mechanical enforcement. Folding it into
`check-unfinished-work.sh` — rejected: it is a different question at a different time, and no
script in this repository answers two.

### `/myflow-start` asks rather than assumes, everywhere approval is sought

**ID:** `planning-gate-always-ask`
**Status:** active
**Chosen:** A new `myflow-planning-gate` capability holding two rules: every unresolved question is
put to the operator rather than resolved by assumption, and every approval or choice a `/myflow-*`
command asks for is offered as options rather than open prose.
**Considered:** Adding it to `myflow-effort` — rejected: that spec's subject is the effort level.
Contract text with no spec — rejected as reproducing exactly the unspecced shape this change is
fixing elsewhere. Scoping the options rule to `/myflow-start` alone, or to `/myflow-finish` alone —
rejected in favour of one rule covering every command, which subsumes both readings of ask 13 and
costs a single sentence.

## Design

### 1. `myflow-jira-projection` — a new capability

`jira-integration.md` is 158 lines of contract with nothing normative behind it. The capability
makes it answerable, holding: the three transition points and when each fires, forward-only
ordering matched by name, the unknown-status ask, labels on created issues, append-only description
sync with its pre-write assertion, and the invariant that Jira is never a gate.

`jira-integration.md` remains the operational contract the commands read; the spec is what it must
satisfy. Neither restates the other.

### 2. The finish run-1 gate

`scripts/check-unfinished-work.sh <worktree> <change-name>` runs **before** the "how should this
branch land?" question and before any git action, once per worktree in the state file's `worktrees`
map.

| Signal | Read from |
|--------|-----------|
| Unticked checklist boxes | `docs/manual-test/<name>.md` |
| Unchecked plan items | `openspec/changes/<name>/tasks.md` and any nested `<name>-fix-N` |
| Findings not closed | `.superpowers/sdd/final-review-panel.md` |
| Work the run knows is undone | the `## Known incomplete` section of the test guide |

Verdicts, following `check-finish-preflight.sh`'s contract: `CLEAR` when nothing is outstanding;
`OUTSTANDING` with a per-signal breakdown otherwise; and **exit 2 with no verdict line** when the
worktree cannot be read, which is treated as stop-and-ask rather than as either verdict. A caller
that greps for `CLEAR` in empty output finds nothing, so the exit code is checked as well as the
line.

**A missing file or a missing `## Known incomplete` section counts as outstanding, not clear.**
Silence is what let the kan-6 case through. Guides written before this change carry no such section
and will therefore prompt once — correct behaviour, not a false positive.

On `OUTSTANDING`, the run shows the breakdown and offers the three answers from
`unfinished-work-prompt`. **File a Jira task** creates an issue carrying the outstanding list,
linked to the change's issue and labelled per `jira-created-issue-labels`, then continues.
**Continue** proceeds and is recorded per `unfinished-work-audit`. **Stop** exits without touching
git, leaving the change at `IN_PROGRESS`.

`/myflow-do` gains the matching duty: write `## Known incomplete` into the guide — either the
literal `None.` or a bullet list — and refresh it on every fix run.

### 3. Findings become countable

`.superpowers/sdd/final-review-panel.md` gains a required table:

| Slot | Severity | Location | Status | Note |
|------|----------|----------|--------|------|

`status` is `open`, `fixed` or `withdrawn`. `/myflow-do` may not hand off while any row is `open`,
which is ask 5. The run-1 gate counts the same rows, so the two surfaces cannot disagree.

### 4. The commit split

`/myflow-do` stages with an exclusion pathspec:

```bash
git -C "$WT" add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
```

Verified by hand on git 2.50.1 in a scratch repository: with `src/a.kt`,
`openspec/changes/x/tasks.md` and `docs/manual-test/x.md` present, that command stages `src/a.kt`
alone and leaves both excluded paths listed by `git ls-files --others --exclude-standard`.

The handoff's review command simplifies to `git -C <worktree> diff --cached`, since there is no
longer anything to filter out. Finish run 1 commits the code, then stages and commits the three
excluded paths as a second commit whose message carries the outstanding list when the operator
chose to continue past the gate. `/myflow-do`'s `prUrl` exception makes the same two commits and
pushes both.

The preserve-session-records call keeps its existing position — before the first `git add -A` in
finish, after staging in `/myflow-do` — because `docs/superpowers/` is one of the excluded paths
and is picked up by the second staging pass either way.

### 5. The cleanup registry

One table in `pipeline.md`, replacing cleanup rules currently scattered across four files, two of
which are silent:

| Artifact | Created by | Location | Removed by |
|----------|-----------|----------|-----------|
| `task-N.diff`, `final-review.diff`, `fix-round-N.diff` | `/myflow-do` | `.superpowers/sdd/` in the worktree | with the worktree, run 2 |
| Panel record | `/myflow-do` | `.superpowers/sdd/` | preserved at run 1; removed with the worktree |
| SDD ledger | `/myflow-do` | `.superpowers/sdd/` | preserved at run 1; removed with the worktree |
| Proposal artifact source | `/myflow-start` | the state directory | run 2, only if a preserved copy exists |
| Worktree | `/myflow-do` | per `worktrees` keys | run 2, after its four checks |
| Local branch | `/myflow-do` | the repository | run 2, `git branch -d` |
| **Remote branch** | finish run 1 | `origin` | **run 2 — new** |
| Change directory | `/myflow-start` | `openspec/changes/<name>/` | moved to the archive, never deleted |
| State file | every command | the state directory | never — it is the terminal record |

`scripts/check-cleanup-complete.sh <repo> <name> <state-dir>` then verifies every row whose
lifetime ends at run 2 and reports anything left behind, rather than assuming removal worked.

### 6. `myflow-planning-gate` — a new capability

Two requirements: unresolved questions are put to the operator rather than resolved by assumption,
and every approval or choice is offered as options rather than open prose. It is the natural home
for later planning asks — slice B's prompts and KAN-13's build-order rule.

## Spec deltas

| Capability | Operation |
|------------|-----------|
| `myflow-jira-projection` | **ADDED** |
| `myflow-planning-gate` | **ADDED** |
| `myflow-finish-cleanup` | **MODIFIED** — the gate before the route question, remote-branch removal, registry verification |
| `myflow-handoff-output` | **MODIFIED** — display filter becomes a staging rule; the two-commit split |
| `myflow-command-surface` | **MODIFIED** — git boundaries follow the split |
| `myflow-review-panel-economics` | **MODIFIED** — the findings table and the zero-open-findings bar |

Files carrying the operational text: `skills/myflow-contracts/pipeline.md`,
`skills/myflow-contracts/jira-integration.md`, `skills/myflow-start/SKILL.md`,
`skills/myflow-do/SKILL.md`, `skills/myflow-finish/SKILL.md`, and `README.md`'s command table.

## Testing

Two new scripts, each with a harness beside it —
`scripts/test-check-unfinished-work.sh` and `scripts/test-check-cleanup-complete.sh`. Four of this
repository's five guard scripts already have one; `check-vocabulary.sh` is the exception, and this
change does not close that gap.

The three existing guards must stay clean: `check-vocabulary.sh`, `check-references.sh`,
`check-plan-provenance.sh`. The last of these now scans this change's own `tasks.md`, `design.md`
and `proposal.md`, so every fenced block and every number in them carries a provenance tag.

The Jira behaviour cannot be unit-tested — there is no Jira CLI on this machine and the integration
is MCP-only. It is verified by the manual-test guide against this change's own issue.

## Risks and trade-offs

- **An abandoned planning run leaves its issue In Progress.** Accepted with
  `jira-in-progress-at-start`; the board being right during a long run was judged worth more.
- **Every severity blocking can cost fix rounds on matters of taste.** Bounded by the panel's
  existing round-5 handback, where the operator adjudicates.
- **Guides predating this change have no `## Known incomplete` section** and will prompt once at
  their next finish. Correct, and self-clearing.
- **Two scripts and two harnesses is the heavier of the two options offered** for the cleanup
  verification. Chosen deliberately for mechanical enforcement.
- **Moving ask 13 into this change widens slice A.** The alternative was slice B amending a spec
  slice A had just written.

## Out of scope

The seven slice-B asks listed under **Scope**, tracked by a follow-up issue filed at the end of this
planning run. Renaming `myflow-review-panel-economics` to match its widened subject. Closing
`check-vocabulary.sh`'s missing harness. KAN-13's planning build-order rule, KAN-15, KAN-16, KAN-24
and KAN-25, which are separately tracked.
