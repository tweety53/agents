# myflow-finish — rationale

This file is the reasoning behind `skills/myflow-finish/SKILL.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

**Load `skills/myflow-contracts/pipeline.md` first** — it is canonical for the states, git
boundaries, the finish contract, and the handoff output shape.

## State gate

## Deciding which run this is

A PR a human merged on the forge, a colleague's merge, and run 1's own merge are indistinguishable
here, and that is correct — all three mean the same thing.

and deliberately does not restate it: a second copy of a procedure that deletes worktrees is a second
copy that can drift.

# Run 1 — integrate

## 1.0 Check for unfinished work

The signals that make a change outstanding are the script's own and are not restated here — a second
list of them would drift from the one that is actually run.

What each verdict means, and what each course below does, is canonical under
**Finish contract** (`skills/myflow-contracts/finish-contract.md`); this section is how it is
executed.

Asking the landing question first and only then reporting unfinished work would make the operator
choose a route for a branch they have not yet been told is incomplete.

The course is labelled for both outcomes because joining is now the
usual one, and an option promising to *file* a task while normally joining an existing issue
describes the wrong write to the operator being asked.

## 1.1 Ask how the branch should land, before any git action

Re-asking cold invites a duplicate PR.

## 1.2 Commit the staged work

**The `reset` is what makes the split hold; the exclusion alone only assumes it** — the reason is
stated under **Git boundaries** (`skills/myflow-contracts/pipeline.md`), which `/myflow-do` follows
too. What is specific to this gate is *whose* staging it retracts: the operator may have run their
own `git add -A` while reviewing, and the excluding `add` cannot take those paths back out.

**A non-zero exit means a copy was attempted and refused or failed:** report
it with the script's own stderr message and continue the integration, per the outcome table under
**Preserving the session records** in `skills/myflow-contracts/pipeline.md`, which is canonical
for all three outcomes.

`scripts/preserve-session-records.sh` still runs **before** the first `add`, unchanged:
`docs/superpowers/` is one of the excluded paths, so its files are picked up by the second staging
pass. The second `add` carries no pathspec, which is what makes it pick them up.

The planning artifacts were hidden from the review diff, not from the commit.

The newest commit on a branch is the one a forge shows first, and that should be the code.

## 1.3 Take the chosen route

## 1.4 No verification gate

## 1.5 State and handoff

The block below is **not** a second definition of the handoff. It is this run's rendering of the
`IN_PROGRESS`-after-run-1 template, which is defined once under
**The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`) and is canonical for the
labels, the field set and their order. What this block adds is the enumeration of the literal
alternatives run 1 writes — the three `PR:` cases below are that enumeration of the template's
placeholder, one per landing route. **Change the template first and bring this block with it** — a
field added here and not there is drift the moment `/myflow-status <name>` regenerates the same
state.

**Outstanding is the same list the planning commit carries**, and it is stated in both places
because either alone is a record the next reader may never reach: a commit message nobody reads
back, or a handoff that scrolls out of a session nobody kept.

The last line is this same command, because that is what the operator runs once the branch is
merged.

The merge-and-push route is why this is a choice at all. It merges into the base branch inside
**1.3**, before this block prints, so a run that took it and then told the operator to *wait on the
merge* would be contradicting the `Route:` line directly beneath the heading, which reads *merged
and pushed*. On the other two routes nothing this run did put the branch onto the base branch, so the
merge is genuinely still ahead.

# Run 2 — archive and clean up

every rule the worktree half carries — the gating checks, the ignored-file disclosure, the
removal sequence, and the remote deletion with its already-gone case — is canonical in
**Worktree cleanup** (`skills/myflow-contracts/finish-contract.md`). Neither is restated here.

Carry `artifactUrl`, `jiraIssue`, `planningEffort`, `models`,
`reviewPanelRoster` and `prUrl` forward — the planning effort as the **mapped level under
`planningEffort`** when the file recorded it under the retired key, per the carry-forward rule in
**State file** (`skills/myflow-contracts/state-file.md`), which is canonical and is not restated
here.

The procedure — skippable per run with running it the default, gathering
input via a script rather than an inline re-read, one combined reasoning pass across all four
angles plus the rating, the per-finding filing ask, and the report path — is canonical under
**Run 2 — the branch is merged** (`skills/myflow-contracts/finish-contract.md`), step 8. The
requirement to change first when that procedure changes is
**Requirement: Self-review runs only after FINISHED is written**
(`openspec/specs/myflow-self-review/spec.md`) — a citation `finish-contract.md` already carries,
not restated here.

Labelling a filed issue and handling a filing failure follow **Labels on issues the pipeline
creates** and **Never blocking** (`skills/myflow-contracts/jira-integration.md`) verbatim — not
restated here.

   The inner `{ commit && push; }` grouping matters: without it, `&&`/`||` at equal precedence
   parse left-to-right as `(diff --quiet || commit) && push`, which runs `push` unconditionally
   once the outer group is entered — even when nothing was staged or committed. Grouping `commit`
   and `push` together means `push` only runs on the branch where `commit` actually ran, matching
   the skip-if-empty shape **Git boundaries** (`skills/myflow-contracts/pipeline.md`) already
   documents for the worktree commits.

A self-review mechanism that itself burned disproportionate tokens on
four separate dispatches would be its own finding under the cost angle — the one-pass shape is
how this step avoids becoming that finding.

## Guardrails
