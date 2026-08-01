# Manual test — kan-36-more-brainstorming-rounds-if-needed

This repository has **no runnable application**. It is the source of the myflow skills, commands and
rules, installed elsewhere by `setup.sh`. There is nothing to start, no port and no URL, so every
check below is either a command to run or a behaviour to exercise by invoking a `/myflow-*` command
in a session.

Run the commands from `/Users/tweety53/Projects/agents`. The behavioural checks need a session in
that repository with the skills installed; the installed copies live under `~/.claude/skills/`, so
run `/Users/tweety53/Projects/agents/setup.sh global` first if you want to exercise them from the
installed tree rather than the working tree.

## Guards and harnesses

- [ ] check `scripts/check-vocabulary.sh` — exits 0
- [ ] check `scripts/check-references.sh` — exits 0, so every citation this change added resolves
- [ ] check `scripts/check-plan-provenance.sh` — exits 0
- [ ] check `scripts/test-check-references.sh` — exits 0
- [ ] check `scripts/test-check-plan-provenance.sh` — exits 0
- [ ] check `HOME="$(mktemp -d)" ./setup.sh global` — installs without touching the real home directory

## myflow-planning-gate — brainstorming converges

- [ ] check `/myflow-start` on a new change — it opens another question round when an answer of
      yours leaves it holding a question, instead of moving on to approaches
- [ ] check that a design section you respond to with a *question* rather than a correction sends it
      back to a question round, not to a design revision
- [ ] check that raising a question at the spec-review gate reopens the loop
- [ ] check that adding scope at any gate is treated as an exchange the convergence test runs after
- [ ] check the confirm — when it holds nothing it states what it believes settled and asks whether
      anything is still unclear, rather than proceeding silently
- [ ] check that naming something at the confirm opens another round
- [ ] check the third round — it is offered as a named choice showing what is still open, not opened
      silently
- [ ] check that rounds one and two are not offered, only taken
- [ ] check that no round count ends the stage on its own — only your answer does
- [ ] check that answering "I cannot answer this" records the question and does **not** reopen a
      round on that same question
- [ ] check `/myflow-start` at `detailed` and at `low` — both still loop; only how many questions each
      round groups differs
- [ ] check a revision round — re-running `/myflow-start` at `STARTED` loops on what your feedback
      reopened and does not re-brainstorm settled parts

## myflow-planning-gate — a question left open is recorded

- [ ] check that choosing *record what is open and move on* writes the question into
      `## Open questions` in the change's `design.md`
- [ ] check the entry carries an immutable ID, a status of `open`, why it is open, and what it affects
- [ ] check that a later round answering the question sets its status to `answered by <decision-id>`
      and adds the decision, leaving the original entry in place
- [ ] check a change that left nothing open — the section is present and empty, not absent

## myflow-handoff-output — the count reaches the gate

- [ ] check the `STARTED` handoff carries `Open questions:` immediately after `Decisions recorded:`
- [ ] check it reads `none` when nothing is open
- [ ] check `/myflow-status <name>` renders the same line — it is regenerated from `design.md`, not
      remembered from the run that printed it
- [ ] check that answering an open question in a revision round lowers the count `/myflow-status`
      shows
- [ ] check the published proposal artifact carries the open questions

## myflow-jira-projection — `TO DO URGENT` is a `To Do` synonym

- [ ] check `/myflow-start` on an issue sitting at `TO DO URGENT` — it transitions to In Progress
      and asks nothing
- [ ] check an issue at a status outside the mapping still raises the one yes/no question
- [ ] check that declining that question leaves the status untouched and reports one line
- [ ] check an issue already at or past the target — no transition is attempted and nothing is asked
- [ ] check `/myflow-finish` run 1's follow-up join search still offers a candidate sitting at
      `TO DO URGENT`

## `/myflow-start` stages and never commits

- [ ] check `/myflow-start` leaves the design document and the change directory **staged and
      uncommitted** — `git log` shows no new commit from the run
- [ ] check it creates no branch and no worktree, and does not push
- [ ] check `/myflow-finish` run 1 still commits the planning artifacts, the test guide and the
      preserved session records in its second commit

## Known incomplete

- **The plan is not visible in `/myflow-do`'s worktree, and this change does not fix it.** A plan
  that is only staged is absent from a worktree built from the base branch, and nothing copies it
  across. This is recorded as an open question in the change's `design.md` and is left for a
  separate change. It is the reason slice D originally tried to make `/myflow-start` commit; that
  attempt was reverted after review, so the gap is unchanged from before this change rather than
  newly introduced.
- **The review panel's Bugbot slot never ran.** The `bugbot` subagent type is not available in this
  harness, so the required defect-hunt slot was dispatched as a general-purpose reviewer on the
  panel's model with a defect-only mandate. It is a substitute, not Bugbot, and carries no Bugbot
  agent definition. The panel record names it as such on both passes.
- **The Security slot was not selected** on either pass, and that is a roster decision rather than an
  omission: no trigger fired — the diff touches no auth, tokens, crypto, secrets, path handling,
  deserialization, HTTP edge or dependencies, and it does not change JQL assembly, only which status
  names map to a position.
- **`tasks.md`'s line-number citations are stale.** Several slices and three fix rounds edited the
  same files, so every line number the plan cites has drifted; each implementer re-located its
  targets by content. The plan is a planning artifact rather than a live index.
- **Nothing was committed by this run.** No `prUrl` is recorded, so the implementation is staged and
  uncommitted, per the git boundary in force.
- **This change was implemented without a worktree**, in the main checkout on branch
  `openspec/kan-36-more-brainstorming-rounds-if-needed`, by explicit operator consent — the plan was
  staged-but-uncommitted, so a worktree branched from `main` would not have contained it. That is the
  same gap recorded as an open question above, encountered during the run.
